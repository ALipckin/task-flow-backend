package port

type Sender interface {
	Send(to, subject, body string) error
}
