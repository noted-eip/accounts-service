package validators

import (
	mailingv1 "accounts-service/protorepo/noted/mailing/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func ValidateSendEmailRequest(in *mailingv1.SendEmailRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.MarkdownBody, validation.Required),
		validation.Field(&in.Subject, validation.Required),
		validation.Field(&in.Recipients, validation.Required, validation.NotNil),
		validation.Field(&in.Recipients, validation.Required),
	)
}
