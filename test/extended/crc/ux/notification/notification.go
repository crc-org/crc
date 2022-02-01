package notification

type Notification interface {
	CheckProcessNotification(process string) error
	ClearNotifications() error
}
