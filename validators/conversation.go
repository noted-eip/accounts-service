package validators

import (
	conversationsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateCreateConversationRequest(in *conversationsv1.CreateConversationRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.Title, validation.Required, validation.Length(1, 20)),
	)
}

func ValidateGetConversationRequest(in *conversationsv1.GetConversationRequest) error {
	return validation.Validate(in.ConversationId,
		validation.Required,
		is.UUID,
	)
}

func ValidateUpdateConversationRequest(in *conversationsv1.UpdateConversationRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.ConversationId, validation.Required, is.UUID),
		validation.Field(&in.Title, validation.Required, validation.Length(1, 20)),
	)
}

func ValidateDeleteConversationRequest(in *conversationsv1.DeleteConversationRequest) error {
	return validation.Validate(in.ConversationId,
		validation.Required,
		is.UUID,
	)
}

func ValidateListConversationRequest(in *conversationsv1.ListConversationsRequest) error {
	return validation.Validate(in.GroupId,
		validation.Required,
		is.UUID,
	)
}
