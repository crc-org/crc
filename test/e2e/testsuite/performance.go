package testsuite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/crc-org/crc/v2/test/extended/util"
	"github.com/shirou/gopsutil/v4/cpu"
)

type Monitor struct {
	cancelFunc context.CancelFunc
	isRunning  bool
	mu         sync.Mutex
	wg         sync.WaitGroup
	interval   time.Duration
}

func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval: interval,
	}
}

func (m *Monitor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("The collector is running")
	}

	fmt.Printf("Attempt to start CPU collector, interval: %s\n", m.interval)

	//  create a context.WithCancel
	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFunc = cancel
	m.isRunning = true

	// start goroutine
	m.wg.Add(1)
	go m.collectLoop(ctx)

	fmt.Println("CPU collector has been successfully started")
	return nil
}

func (m *Monitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return fmt.Errorf("The collector is not running")
	}
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	m.isRunning = false
	m.wg.Wait()
	fmt.Println("CPU collector has sent a stop signal")
	// may need wait a while to stop
	return nil
}

func (m *Monitor) collectLoop(ctx context.Context) {
	defer m.wg.Done()

	fmt.Println("--> collect goroutine start...")
	calcInterval := m.interval

	// Prime cpu.Percent so the subsequent calls with interval=0 return usage
	// since this priming call. We discard the returned value.
	_, _ = cpu.Percent(0, false)

	ticker := time.NewTicker(calcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("<-- collect goroutine receive stop signal")
			return // exit goroutine
		case <-ticker.C:
			// cpu.Percent with interval=0 returns percent since the last call.
			totalPercent, err := cpu.Percent(0, false)
			if err != nil {
				fmt.Printf("Error: fail to collect CPU data: %v\n", err)
				// do not block here; wait for next tick or cancellation
				continue
			}

			if len(totalPercent) > 0 {
				data := fmt.Sprintf("[%s], cpu percent: %.2f%%\n",
					time.Now().Format("15:04:05"), totalPercent[0])
				wd, err := os.Getwd()
				if err != nil {
					fmt.Printf("Error: failed to get working directory: %v\n", err)
					continue
				}
				file := filepath.Join(wd, "../test-results/cpu-consume.txt")
				err = util.WriteToFile(data, file)
				if err != nil {
					fmt.Printf("Error: fail to write to %s: %v\n", file, err)
				}
			}
		}
	}
}
