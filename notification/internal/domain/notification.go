package domain

type User struct {
	ID    int
	Email string
	Name  string
}

type Notification struct {
	UserID  int
	Email   string
	Event   string
	Payload string
	SentAt  string
	Channel string
}

type NotificationPayload struct {
	Event       string `json:"event"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type OutboundMessage struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}
