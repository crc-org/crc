package machine

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestOneStartAtTheSameTime(t *testing.T) {
	isRunning := make(chan struct{}, 1)
	startCh := make(chan struct{}, 1)
	waitingMachine := &waitingMachine{
		isRunning:       isRunning,
		startCompleteCh: startCh,
	}
	syncMachine := NewSynchronizedMachine(waitingMachine)
	assert.Equal(t, Idle, syncMachine.CurrentState())

	lock := &sync.WaitGroup{}
	lock.Add(1)
	go func() {
		defer lock.Done()
		_, err := syncMachine.Start(context.Background(), types.StartConfig{})
		assert.NoError(t, err)
	}()

	<-isRunning
	assert.Equal(t, Starting, syncMachine.CurrentState())
	assert.Equal(t, waitingMachine.GetName(), syncMachine.GetName())
	_, err := syncMachine.Start(context.Background(), types.StartConfig{})
	assert.EqualError(t, err, "cluster is busy")

	startCh <- struct{}{}
	lock.Wait()

	assert.Equal(t, Idle, syncMachine.CurrentState())
}

func TestDeleteStop(t *testing.T) {
	isRunning := make(chan struct{}, 1)
	deleteCh := make(chan struct{}, 1)
	waitingMachine := &waitingMachine{
		isRunning:        isRunning,
		deleteCompleteCh: deleteCh,
	}
	syncMachine := NewSynchronizedMachine(waitingMachine)
	assert.Equal(t, Idle, syncMachine.CurrentState())

	lock := &sync.WaitGroup{}
	lock.Add(1)
	go func() {
		defer lock.Done()
		assert.NoError(t, syncMachine.Delete())
	}()

	<-isRunning
	assert.Equal(t, Deleting, syncMachine.CurrentState())
	assert.EqualError(t, syncMachine.Delete(), "cluster is stopping or deleting")
	_, err := syncMachine.Stop()
	assert.EqualError(t, err, "cluster is stopping or deleting")
	_, err = syncMachine.Start(context.Background(), types.StartConfig{})
	assert.EqualError(t, err, "cluster is busy")

	deleteCh <- struct{}{}
	lock.Wait()

	assert.Equal(t, Idle, syncMachine.CurrentState())
}

func TestCancelStart(t *testing.T) {
	isRunning := make(chan struct{}, 1)
	deleteCh := make(chan struct{}, 1)
	waitingMachine := &waitingMachine{
		isRunning:        isRunning,
		startCompleteCh:  make(chan struct{}, 1),
		deleteCompleteCh: deleteCh,
	}
	syncMachine := NewSynchronizedMachine(waitingMachine)
	assert.Equal(t, Idle, syncMachine.CurrentState())

	lock := &sync.WaitGroup{}
	lock.Add(1)
	go func() {
		defer lock.Done()
		_, err := syncMachine.Start(context.Background(), types.StartConfig{})
		assert.EqualError(t, err, "context canceled")
	}()

	<-isRunning
	assert.Equal(t, Starting, syncMachine.CurrentState())

	lock.Add(1)
	go func() {
		defer lock.Done()
		assert.NoError(t, syncMachine.Delete())
	}()

	deleteCh <- struct{}{}
	lock.Wait()

	assert.Equal(t, Idle, syncMachine.CurrentState())
}

type waitingMachine struct {
	isRunning        chan struct{}
	startCompleteCh  chan struct{}
	stopCompleteCh   chan struct{}
	deleteCompleteCh chan struct{}
}

func (m *waitingMachine) IsRunning() (bool, error) {
	return false, errors.New("not implemented")
}

func (m *waitingMachine) GetName() string {
	return "waiting machine"
}

func (m *waitingMachine) Delete() error {
	m.isRunning <- struct{}{}
	<-m.deleteCompleteCh
	return nil
}

func (m *waitingMachine) Exists() (bool, error) {
	return false, errors.New("not implemented")
}

func (m *waitingMachine) GetConsoleURL() (*types.ConsoleResult, error) {
	return nil, errors.New("not implemented")
}

func (m *waitingMachine) IP() (string, error) {
	return "", errors.New("not implemented")
}

func (m *waitingMachine) PowerOff() error {
	return nil
}

func (m *waitingMachine) Start(context context.Context, _ types.StartConfig) (*types.StartResult, error) {
	m.isRunning <- struct{}{}
	select {
	case <-context.Done():
		return nil, context.Err()
	case <-m.startCompleteCh:
		return &types.StartResult{
			Status:         state.Running,
			KubeletStarted: true,
		}, nil
	}
}

func (m *waitingMachine) Status() (*types.ClusterStatusResult, error) {
	return nil, errors.New("not implemented")
}

func (m *waitingMachine) Stop() (state.State, error) {
	m.isRunning <- struct{}{}
	<-m.stopCompleteCh
	return state.Stopped, nil
}
