package domain

type User struct {
	ID    int
	Email string
	Name  string
}

type NotificationPayload struct {
	Event       string
	Title       string
	Description string
}

type OutboundMessage struct {
	UserID  int
	Email   string
	Message string
}
