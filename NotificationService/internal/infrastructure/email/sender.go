package email

import (
	"NotificationService/internal/config"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// Sender интерфейс для отправки email
type Sender interface {
	Send(to, subject, body string) error
}

// SMTPEmailSender — реализация отправки email через SMTP
type SMTPEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	auth     smtp.Auth
}

// NewSMTPEmailSender конструктор
func NewSMTPEmailSender(cfg *config.SMTPConfig) Sender {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &SMTPEmailSender{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		auth:     auth,
	}
}

// Send отправляет email сообщение
func (s *SMTPEmailSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	msg := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n" +
		body

	// TLS конфигурация (если нужен)
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(s.auth); err != nil {
		return err
	}

	if err = client.Mail(s.username); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
