package main

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"fmt"

	mailing "github.com/noted-eip/noted/mailing-service"
)

func ForgetAccountPasswordMailContent(accountID string, token string) *mailing.SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour,<br/>Nous avons reçu une demande pour réinitialiser votre mot de passe.
		<br/>Si vous n'avez pas fait la demande, ignorez simplement ce message.
		<br/>Sinon, vous pouvez réinitialiser votre mot de passe.
		<br/>Attention, votre code n'est valable qu'une heure.
		<br/><div style="padding:16px 24px;border:1px solid #eeeeee;background-color:#f4f4f4;
		border-radius:3px;font-family:monospace;margin:24px 0px 24px 0px ">%s</div></span>`, token)

	return &mailing.SendEmailsRequest{
		To:      []string{accountID},
		Sender:  "noted.organisation@gmail.com",
		Title:   "Mise à jour de mot de passe",
		Subject: "Réinitialisez votre mot de passe",
		Body:    body,
	}
}

func ValidateAccountByEmail(accountID string, token string) *mailing.SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour,<br/>Voici votre code de validation Noted.
		<br/>Si vous n'avez pas fait la demande, ignorez simplement ce message.
		<br/><div style="padding:16px 24px;border:1px solid #eeeeee;background-color:#f4f4f4;
		border-radius:3px;font-family:monospace;margin:24px 0px 24px 0px ">%s</div></span>`, token)

	return &mailing.SendEmailsRequest{
		To:      []string{accountID},
		Sender:  "noted.organisation@gmail.com",
		Title:   "Noted: Code de validation",
		Subject: "Valider votre compte Noted",
		Body:    body,
	}
}

func SendGroupInviteMailContent(in *accountsv1.SendGroupInviteMailRequest) *mailing.SendEmailsRequest {
	body := fmt.Sprintf(`<span>Bonjour,<br/>Vous avez été invité à rejoindre le groupe %s.
	<br/>Veuillez vous connecter à votre profil pour accepter ou refuser l'invitation.
	<a href="https://noted-eip.vercel.app/profile" style="color: blue">Visitez mon profil (https://noted-eip.vercel.app/profile) </a>
	<br/>Attention, cette invitation est valable seulement 2 semaines</span>`, in.GroupName)

	return &mailing.SendEmailsRequest{
		To:      []string{in.RecipientId},
		Sender:  "noted.organisation@gmail.com",
		Title:   "Invitation à rejoindre un groupe",
		Subject: "Vous avez été invité à rejoindre le groupe " + in.GroupName,
		Body:    body,
	}
}
