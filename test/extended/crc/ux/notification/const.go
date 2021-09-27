package notification

const (
	startMessage  string = "OpenShift cluster is running"
	stopMessage   string = "The OpenShift Cluster was successfully stopped"
	deleteMessage string = "The OpenShift Cluster is successfully deleted"
	// copyCommandMessage string = "OC Login command copied to clipboard, go ahead and login to your cluster"

	notificationWaitTimeout int = 200
	notificationWaitRetries int = 10
)
