package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateDeleteGroupRequest(in *accountsv1.DeleteGroupRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
	)
}

func ValidateUpdatedGroupRequest(in *accountsv1.UpdateGroupRequest) error {
	return validation.ValidateStruct(in.Group,
		validation.Field(&in.Group.Id, validation.Required, is.UUID),
	)
}

func ValidateListGroups(in *accountsv1.ListGroupsRequest) error {
	return validation.Validate(&in.AccountId, validation.Required, is.UUID)
}
