package notification

type Notification interface {
	GetClusterRunning() error
	GetClusterStopped() error
	GetClusterDeleted() error
	ClearNotifications() error
}
