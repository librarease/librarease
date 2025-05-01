package usecase

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
