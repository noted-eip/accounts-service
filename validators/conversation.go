package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateCreateConversationRequest(in *accountsv1.CreateConversationRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.Title, validation.Required, validation.Length(1, 30)),
	)
}

func ValidateGetConversationRequest(in *accountsv1.GetConversationRequest) error {
	return validation.Validate(in.ConversationId,
		validation.Required,
		is.UUID,
	)
}

func ValidateUpdateConversationRequest(in *accountsv1.UpdateConversationRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
		validation.Field(&in.Title, validation.Required, validation.Length(1, 30)),
	)
}

func ValidateDeleteConversationRequest(in *accountsv1.DeleteConversationRequest) error {
	return validation.Validate(in.ConversationId,
		validation.Required,
		is.UUID,
	)
}

func ValidateListConversationRequest(in *accountsv1.ListConversationsRequest) error {
	return validation.Validate(in.GroupId,
		validation.Required,
		is.UUID,
	)
}
