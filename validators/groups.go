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
	err := validation.ValidateStruct(in,
		validation.Field(&in.AccountId, validation.Required, is.UUID),
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

func ValidateGetGroup(in *accountsv1.GetGroupRequest) error {
	return validation.Validate(&in.GroupId, validation.Required, is.UUID)
}
