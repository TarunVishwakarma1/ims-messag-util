// Notification Service for Web App. Whenever a notification is triggered from the web app, it will be sent to the user via this service.
// Or when a service triggers a notification it will be triggered via this service.
package notification

import "fmt"

type Notification interface {
	SendNotification()
}

type Message struct {
	UserId           string
	Message          string
	NotificationType string
}

type Alert struct {
	Message Message
}

type Information struct {
	Message Message
}

type Warning struct {
	Message Message
}

func (a *Alert) SendNotification() {
	fmt.Println("Alert: ", a.Message.Message)
}

func (i *Information) SendNotification() {
	fmt.Println("Information: ", i.Message.Message)
}

func (w *Warning) SendNotification() {
	fmt.Println("Warning: ", w.Message.Message)
}

func SendNotification(n Notification) {
	n.SendNotification()
}
