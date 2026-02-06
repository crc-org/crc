package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	log "github.com/sirupsen/logrus"
)

type NotificationSender struct {
	notificationCh chan types.NotificationMessage
	socket         string
}

func NewNotificationSender(socket string) *NotificationSender {
	if socket == "" {
		return &NotificationSender{
			socket:         "",
			notificationCh: nil,
		}
	}

	return &NotificationSender{
		socket:         socket,
		notificationCh: make(chan types.NotificationMessage, 100),
	}
}

func (s *NotificationSender) Send(notification types.NotificationMessage) {
	if s.notificationCh == nil {
		return
	}
	select {
	case s.notificationCh <- notification:
	default:
		log.Warn("unable to send notification")
	}
}

func (s *NotificationSender) Start(ctx context.Context) {
	if s.notificationCh == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-s.notificationCh:
			if err := s.sendToSocket(notification); err != nil {
				log.Errorf("failed to send notification: %v", err)
				continue
			}
		}
	}
}

func (s *NotificationSender) sendToSocket(notification types.NotificationMessage) error {
	if s.socket == "" {
		return nil
	}
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: s.socket, Net: "unix"})
	if err != nil {
		return fmt.Errorf("cannot dial notification socket: %w", err)
	}
	defer conn.Close()
	enc := json.NewEncoder(conn)
	if err := enc.Encode(notification); err != nil {
		return fmt.Errorf("failed to encode notification: %w", err)
	}
	return nil
}
