package email

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/librarease/librarease/internal/usecase"
	"github.com/wneessen/go-mail"
)

func NewEmailProvider(
	smtpHost, smtpUser, smtpPassword, smtpPort string) *EmailProvider {

	if smtpHost == "" || smtpUser == "" || smtpPassword == "" || smtpPort == "" {
		panic("email: SMTP host, user, and password must be provided")
	}

	var (
		smtpPortInt int
		err         error
	)
	if smtpPortInt, err = strconv.Atoi(smtpPort); err != nil {
		panic("email: invalid SMTP port: " + err.Error())
	}

	client, err := mail.NewClient(
		smtpHost,
		mail.WithPort(smtpPortInt),
		mail.WithUsername(smtpUser),
		mail.WithPassword(smtpPassword),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
	)
	if err != nil {
		panic("email: failed to create SMTP client: " + err.Error())
	}

	// Create a channel to send emails
	emailChan := make(chan *mail.Msg, 100)

	provider := &EmailProvider{
		c:      emailChan,
		client: client,
	}

	// Start a worker to send emails
	go provider.sendEmailWorker()

	return provider
}

type EmailProvider struct {
	c      chan *mail.Msg
	client *mail.Client
}

func (e *EmailProvider) SendEmail(_ context.Context, email usecase.Email) error {
	msg := mail.NewMsg()
	msg.From(email.From)
	msg.To(email.To...)
	msg.Cc(email.CC...)
	msg.Bcc(email.BCC...)
	msg.Subject(email.Subject)
	msg.SetBodyString(mail.TypeTextHTML, email.Body)
	for _, file := range email.Attachments {
		if err := msg.AttachReader(
			file.Name,
			bytes.NewReader(file.Content),
			mail.WithFileContentType(mail.ContentType(file.ContentType)),
		); err != nil {
			fmt.Printf("email: failed to attach file: %v\n", err)
		}
	}

	e.c <- msg

	return nil
}

func (e *EmailProvider) sendEmailWorker() {
	for msg := range e.c {
		if err := e.client.DialAndSend(msg); err != nil {
			fmt.Printf("email: failed to send email: %v\n", err)
		}
	}
}
