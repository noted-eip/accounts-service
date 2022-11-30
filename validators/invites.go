package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateSendInvite(in *accountsv1.SendInviteRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.RecipientAccountId, validation.Required, is.UUID),
		validation.Field(&in.GroupId, validation.Required, is.UUID),
	)
}

func ValidateListInvites(in *accountsv1.ListInvitesRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.SenderAccountId, validation.Required, is.UUID),
		validation.Field(&in.RecipientAccountId, validation.Required, is.UUID),
		validation.Field(&in.GroupId, validation.Required, is.UUID),
	)
}

func ValidateGetInvite(in *accountsv1.GetInviteRequest) error {
	return validation.Validate(&in.InviteId, validation.Required, is.UUID)
}
