package usecase

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type Email struct {
	To          []string
	From        string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments []EmailAttachment
}

type EmailAttachment struct {
	Name        string
	ContentType string
	Content     []byte
}

func (u Usecase) SendBorrowingEmail(ctx context.Context, id uuid.UUID) error {

	b, err := u.repo.GetBorrowingByID(ctx, id, BorrowingsOption{})
	if err != nil {
		return err
	}

	body, err := u.buildBorrowingEmailBody(b)
	if err != nil {
		return err
	}

	email := Email{
		To:      []string{b.Subscription.User.Email},
		From:    "no-reply@librarease.org",
		Subject: "Borrow Confirmation",
		Body:    body,
	}

	return u.mailer.SendEmail(ctx, email)
}

func (u Usecase) buildBorrowingEmailData(b Borrowing) BorrowingEmailData {

	png, _ := qrcode.Encode(b.ID.String(), qrcode.Low, 128)
	png64 := base64.StdEncoding.EncodeToString(png)
	qrCodeURL := "data:image/png;base64," + png64

	return BorrowingEmailData{
		Title:          "Borrowing Confirmation",
		URL:            "https://librarease.org",
		CurrentYear:    time.Now().Format("2006"),
		LibraryName:    b.Subscription.Membership.Library.Name,
		LibraryAddress: b.Subscription.Membership.Library.Address,
		LibraryEmail:   b.Subscription.Membership.Library.Email,
		LibraryPhone:   b.Subscription.Membership.Library.Phone,
		UserName:       b.Subscription.User.Name,
		UserEmail:      b.Subscription.User.Email,
		BookName:       b.Book.Title,
		BookCode:       b.Book.Code,
		MembershipName: b.Subscription.Membership.Name,
		FinePerDay:     b.Subscription.FinePerDay,
		BorrowingID:    b.ID.String(),
		BorrowedAt:     b.BorrowedAt.Format("2006-01-02 03:04 PM"),
		DueAt:          b.DueAt.Format("2006-01-02 03:04 PM"),
		QRCodeURL:      qrCodeURL,
	}
}

//go:embed templates/*
var templates embed.FS

func (u Usecase) buildBorrowingEmailBody(b Borrowing) (string, error) {

	tmpl, err := template.
		New("base.html").
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML {
				return template.HTML(s)
			},
			"safeURL": func(s string) template.URL {
				return template.URL(s)
			},
		}).
		ParseFS(
			templates,
			"templates/base.html",
			"templates/borrowing.html",
		)

	if err != nil {
		return "", err
	}

	data := u.buildBorrowingEmailData(b)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type BorrowingEmailData struct {
	Title       string
	URL         string
	CurrentYear string

	// library
	LibraryName    string
	LibraryAddress string
	LibraryEmail   string
	LibraryPhone   string

	// user
	UserName  string
	UserEmail string

	// book
	BookName string
	BookCode string

	// membership
	MembershipName string

	// subscription
	FinePerDay int

	// borrowing
	BorrowingID string
	BorrowedAt  string
	DueAt       string
	QRCodeURL   string
}
