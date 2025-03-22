package services

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	SMTPHost := os.Getenv("SMTP_HOST")
	SMTPPort := os.Getenv("SMTP_PORT")
	SenderEmail := os.Getenv("SENDER_EMAIL")
	SenderPass := os.Getenv("SENDER_PASSWORD")

	// Создаем объект email
	e := email.NewEmail()
	e.From = SenderEmail
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	a := smtp.CRAMMD5Auth(SenderEmail, SenderPass)
	// Отправка без TLS
	err := e.Send(SMTPHost+":"+SMTPPort, a)
	if err != nil {
		return err
	}

	fmt.Println("Письмо успешно отправлено!")
	return nil
}
