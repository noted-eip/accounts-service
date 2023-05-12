package main

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"fmt"
)

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

func SendGroupInviteMailContent(in *accountsv1.SendGroupInviteMailRequest) *SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour,<br/>Vous avez été invité à rejoindre le groupe %s.
	<br/>Veuillez vous connecter afin d'accepter ou refuser l'invitation.
	<br/>Attention, cette invitation est valable jusqu'au %s</span>`, in.GroupName, in.ValidUntil)

	return &SendEmailsRequest{
		to:      []string{in.RecipientId},
		sender:  "noted.organisation@gmail.com",
		title:   "Invitation à rejoindre un groupe",
		subject: "Vous avez été invité à rejoindre le groupe " + in.GroupName,
		body:    body,
	}
}
