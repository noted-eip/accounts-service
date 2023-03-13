package main

import (
	"accounts-service/models"
	mailingv1 "accounts-service/protorepo/noted/mailing/v1"
	"accounts-service/validators"
	"bytes"
	"context"
	"html/template"
	"net/smtp"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

type mailingAPI struct {
	mailingv1.UnimplementedMailingAPIServer

	logger *zap.Logger
	repo   models.AccountsRepository
}

var _ mailingv1.MailingAPIServer = &mailingAPI{}

type TemplateData struct {
	CODE    string
	CONTENT template.HTML
	TITLE   string
}

type Request struct {
	from    string
	to      []string
	subject string
	body    string
	super   []byte
}

func NewRequest(from string, to []string, subject, body string) *Request {
	return &Request{
		from:    from,
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (r *Request) SendEmailToAccounts() error {
	subject := "Subject: " + r.subject + "!\n"
	mine := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + mine + r.body)
	addr := "smtp.gmail.com:587"
	ssPassword := os.Getenv("GMAIL_SUPER_SECRET")

	if ssPassword == "" {
		return status.Error(codes.Internal, "could not retrieve super secret from environement")
	}

	auth := smtp.PlainAuth("", r.from, string(ssPassword), "smtp.gmail.com")

	if err := smtp.SendMail(addr, auth, r.from, r.to, msg); err != nil {
		return err
	}
	return nil
}

func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

func (srv *mailingAPI) SendEmails(ctx context.Context, in *mailingv1.SendEmailsRequest) (*mailingv1.SendEmailsResponse, error) {

	err := validators.ValidateSendEmailsRequest(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	filters := []*models.OneAccountFilter{}

	for _, val := range in.Recipients {
		filters = append(filters, &models.OneAccountFilter{ID: val.AccountId})
	}

	mails, err := srv.repo.GetMailsFromIDs(ctx, filters)
	if err != nil {
		return nil, statusFromModelError(err)
	}

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	md := []byte(in.MarkdownBody)
	html := markdown.ToHTML(md, parser, nil)

	content := template.HTML(string(html[:]))

	templateData := TemplateData{
		CONTENT: content,
		TITLE:   in.Title,
	}

	r := NewRequest("noted.organisation@gmail.com", mails, in.Subject, in.MarkdownBody)
	err = r.ParseTemplate("ressources/mail.html", templateData)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = r.SendEmailToAccounts()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &mailingv1.SendEmailsResponse{}, nil
}
