package telemetry

import (
	"context"
	"os/user"
	"strings"
	"sync"
)

type contextKey struct{}

var key = contextKey{}

type Properties struct {
	lock    sync.Mutex
	storage map[string]interface{}
}

func (p *Properties) set(name string, value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.storage[name] = value
}

func (p *Properties) values() map[string]interface{} {
	p.lock.Lock()
	defer p.lock.Unlock()
	ret := make(map[string]interface{})
	for k, v := range p.storage {
		ret[k] = v
	}
	return ret
}

func propertiesFromContext(ctx context.Context) *Properties {
	value := ctx.Value(key)
	if cast, ok := value.(*Properties); ok {
		return cast
	}
	return nil
}

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, &Properties{storage: make(map[string]interface{})})
}

func GetContextProperties(ctx context.Context) map[string]interface{} {
	properties := propertiesFromContext(ctx)
	if properties == nil {
		return make(map[string]interface{})
	}
	return properties.values()
}

func setContextProperty(ctx context.Context, key string, value interface{}) {
	properties := propertiesFromContext(ctx)
	if properties != nil {
		properties.set(key, value)
	}
}

func SetCPUs(ctx context.Context, value int) {
	setContextProperty(ctx, "cpus", value)
}

func SetMemory(ctx context.Context, value uint64) {
	setContextProperty(ctx, "memory", value)
}

func SetDiskSize(ctx context.Context, value uint64) {
	setContextProperty(ctx, "disk-size", value)
}

func SetConfigurationKey(ctx context.Context, value string) {
	setContextProperty(ctx, "key", value)
}

func SetStartType(ctx context.Context, value StartType) {
	setContextProperty(ctx, "start-type", value)
}

type StartType string

const (
	AlreadyRunningStartType StartType = "already-running"
	CreationStartType       StartType = "creation"
	StartStartType          StartType = "start"
)

func SetError(err error) string {
	// Mask username if present in the error string
	user, err1 := user.Current()
	if err1 != nil {
		return err1.Error()
	}
	return strings.ReplaceAll(err.Error(), user.Username, "XXXXX")
}
