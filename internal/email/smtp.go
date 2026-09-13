package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	Host, Port, Username, Password, From string
}

func NewSMTPSender(host, port, user, pass, from string) *SMTPSender {
	return &SMTPSender{Host: host, Port: port, Username: user, Password: pass, From: from}
}

func (s *SMTPSender) Send(_ context.Context, msg Message) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n",
		s.From, msg.To, msg.Subject,
	)
	return smtp.SendMail(s.Host+":"+s.Port, auth, s.From, []string{msg.To}, []byte(headers+msg.Body))
}
