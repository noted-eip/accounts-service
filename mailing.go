package main

import (
	"accounts-service/models"
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
	logger      *zap.Logger
	repo        models.AccountsRepository
	pendingRepo models.PendingAccountsRepository
	secret      string
}

type TemplateData struct {
	CODE    string
	CONTENT template.HTML
	TITLE   string
}

func (r *SendEmailsRequest) PostEmails(secret string) error {

	subject := "Subject: " + r.subject + "!\n"
	mine := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + mine + r.body)
	addr := "smtp.gmail.com:587"

	auth := smtp.PlainAuth("", r.sender, secret, "smtp.gmail.com")
	if err := smtp.SendMail(addr, auth, r.sender, r.to, msg); err != nil {
		return err
	}
	return nil
}

func (r *SendEmailsRequest) FormatEmails(templateFileName string) error {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)
	md := []byte(r.body)
	html := markdown.ToHTML(md, parser, nil)
	content := template.HTML(string(html[:]))

	templateData := TemplateData{
		CONTENT: content,
		TITLE:   r.title,
	}

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, templateData)
	if err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

func (srv *mailingAPI) SendEmails(ctx context.Context, req *SendEmailsRequest, isAccountValidation bool) error {

	if srv.secret == "" {
		return status.Error(codes.Internal, "could not retrieve super secret from environement")
	}

	filters := []*models.OneAccountFilter{}
	for _, accountID := range req.to {
		filters = append(filters, &models.OneAccountFilter{ID: accountID})
	}

	var err error
	if !isAccountValidation {
		req.to, err = srv.repo.GetMailsFromIDs(ctx, filters)
	} else {
		req.to, err = srv.pendingRepo.GetMailsFromIDs(ctx, filters)
	}
	if err != nil {
		return statusFromModelError(err)
	}

	err = req.FormatEmails("mail.html")
	if err != nil {
		currentPath, err := os.Executable()
		if err == nil {
			srv.logger.Error("Formating email error from direcory " + currentPath)
		}
		return status.Error(codes.Internal, err.Error())
	}

	err = req.PostEmails(srv.secret)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}
