package services

import (
	"NotifyService/logger"
	"github.com/jordan-wright/email"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	logger.Log(logger.LevelInfo, "Sending email", map[string]any{
		"to":      to,
		"subject": subject,
	})
	SMTPHost := os.Getenv("SMTP_HOST")
	SMTPPort := os.Getenv("SMTP_PORT")
	SenderEmail := os.Getenv("SENDER_EMAIL")
	SenderPass := os.Getenv("SENDER_PASSWORD")

	e := email.NewEmail()
	e.From = SenderEmail
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	a := smtp.CRAMMD5Auth(SenderEmail, SenderPass)
	err := e.Send(SMTPHost+":"+SMTPPort, a)
	if err != nil {
		return err
	}

	return nil
}
