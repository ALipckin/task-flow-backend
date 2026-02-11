package email

import (
	"net/smtp"
	"notification/internal/port"

	"github.com/jordan-wright/email"
)

type SMTPConfig struct {
	Host     string
	Port     string
	From     string
	Password string
}

type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	e := email.NewEmail()
	e.From = s.cfg.From
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	auth := smtp.CRAMMD5Auth(s.cfg.From, s.cfg.Password)
	return e.Send(s.cfg.Host+":"+s.cfg.Port, auth)
}

var _ port.Sender = (*SMTPSender)(nil)
