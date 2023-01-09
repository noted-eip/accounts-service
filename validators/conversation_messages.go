package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateSendConversationMessageRequest(in *accountsv1.SendConversationMessageRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
		validation.Field(&in.Content, validation.Required, validation.Length(1, 250)),
	)
}

func ValidateDeleteConversationMessageRequest(in *accountsv1.DeleteConversationMessageRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.MessageId, validation.Required, is.UUID),
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
	)
}

func ValidateUpdateConversationMessageRequest(in *accountsv1.UpdateConversationMessageRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.MessageId, validation.Required, is.UUID),
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
		validation.Field(&in.Content, validation.Required, validation.Length(1, 250)),
	)
}

func ValidateGetConversationMessageRequest(in *accountsv1.GetConversationMessageRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.MessageId, validation.Required, is.UUID),
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
	)
}

func ValidateListConversationMessageRequest(in *accountsv1.ListConversationMessagesRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
	)
	if err != nil {
		return err
	}
	err = validation.Validate(in.Limit, validation.When(in.Limit != 0, validation.Required), validation.Min(0))
	if err != nil {
		return err
	}
	err = validation.Validate(in.Offset, validation.When(in.Offset != 0, validation.Required), validation.Min(0))
	if err != nil {
		return err
	}
	return nil
}
