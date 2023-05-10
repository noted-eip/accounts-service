package main

import "fmt"

type SendEmailsRequest struct {
	to      []string
	sender  string
	subject string
	title   string
	body    string
}

func ForgetAccountPasswordMailContent(accountID string, token string) *SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour,<br/>Nous avons reçu une demande pour réinitialiser votre mot de passe.
		<br/>Si vous n'avez pas fait la demande, ignorez simplement ce message.
		<br/>Sinon, vous pouvez réinitialiser votre mot de passe.
		<br/>Attention, votre code n'est valable qu'une heure.
		<br/><div style="padding:16px 24px;border:1px solid #eeeeee;background-color:#f4f4f4;
		border-radius:3px;font-family:monospace;margin:24px 0px 24px 0px ">%s</div></span>`, token)

	return &SendEmailsRequest{
		to:      []string{accountID},
		sender:  "noted.organisation@gmail.com",
		title:   "Mise à jour de mot de passe",
		subject: "Réinitialisez votre mot de passe",
		body:    body,
	}
}

func ValidateAccountMailContent(accountID string, name string, token string) *SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour %s,<br/>Vous êtes presque prêt à parcourir Noted.
		<br/>Cliquez sur le lien ci-dessous pour vérifier votre adresse e-mail.
		<br/><div style="padding:16px 24px;border:1px solid #eeeeee;background-color:#f4f4f4;
		border-radius:3px;font-family:monospace;margin:24px 0px 24px 0px ">%s</div></span>`, name, token)

	return &SendEmailsRequest{
		to:      []string{accountID},
		sender:  "noted.organisation@gmail.com",
		title:   "Validation de votre compte",
		subject: "Valider votre compte Noted",
		body:    body,
	}
}
