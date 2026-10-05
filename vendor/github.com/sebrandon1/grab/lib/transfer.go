package lib

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type gauge interface {
	Sample(t time.Time, n int64)
	BPS() float64
}

// bpsGauge computes bytes-per-second over a sliding time window.
type bpsGauge struct {
	mu      sync.Mutex
	samples []bpsSample
	window  time.Duration
}

type bpsSample struct {
	t time.Time
	n int64
}

func newBPSGauge(window time.Duration) *bpsGauge {
	return &bpsGauge{window: window}
}

func (g *bpsGauge) Sample(t time.Time, n int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.samples = append(g.samples, bpsSample{t: t, n: n})
	cutoff := t.Add(-g.window)
	for len(g.samples) > 1 && g.samples[0].t.Before(cutoff) {
		g.samples = g.samples[1:]
	}
}

func (g *bpsGauge) BPS() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.samples) < 2 {
		return 0
	}
	first := g.samples[0]
	last := g.samples[len(g.samples)-1]
	elapsed := last.t.Sub(first.t).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(last.n-first.n) / elapsed
}

type transfer struct {
	n     int64 // must be 64bit aligned on 386
	ctx   context.Context
	gauge gauge
	lim   RateLimiter
	w     io.Writer
	r     io.Reader
	b     []byte
}

func newTransfer(ctx context.Context, lim RateLimiter, dst io.Writer, src io.Reader, buf []byte) *transfer {
	return &transfer{
		ctx:   ctx,
		gauge: newBPSGauge(5 * time.Second),
		lim:   lim,
		w:     dst,
		r:     src,
		b:     buf,
	}
}

// copy behaves similarly to io.CopyBuffer except that it checks for cancelation
// of the given context.Context, reports progress in a thread-safe manner and
// tracks the transfer rate.
func (c *transfer) copy() (written int64, err error) {
	// start the transfer
	if c.b == nil {
		c.b = make([]byte, 32*1024)
	}
	for {
		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		default:
			// keep working
		}
		nr, er := c.r.Read(c.b)
		if nr > 0 {
			nw, ew := c.w.Write(c.b[0:nr])
			if nw > 0 {
				written += int64(nw)
				atomic.StoreInt64(&c.n, written)
				if c.gauge != nil {
					c.gauge.Sample(time.Now(), written)
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
			// wait for rate limiter
			if c.lim != nil {
				err = c.lim.WaitN(c.ctx, nr)
				if err != nil {
					return
				}
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// N returns the number of bytes transferred.
func (c *transfer) N() (n int64) {
	if c == nil {
		return 0
	}
	n = atomic.LoadInt64(&c.n)
	return
}

// BPS returns the current bytes per second transfer rate using a simple moving
// average.
func (c *transfer) BPS() (bps float64) {
	if c == nil || c.gauge == nil {
		return 0
	}
	return c.gauge.BPS()
}
