package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateAddGroupMember(in *accountsv1.AddGroupMemberRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.AccountId, validation.Required, is.UUID),
	)
}

func ValidateRemoveGroupMember(in *accountsv1.RemoveGroupMemberRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.AccountId, validation.Required, is.UUID),
	)
}

func ValidateUpdateGroupMember(in *accountsv1.UpdateGroupMemberRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.Member, validation.Required, is.UUID),
	)
}

func ValidateListGroupMember(in *accountsv1.ListGroupMembersRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
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

func ValidateGetGroupMember(in *accountsv1.GetGroupMemberRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.AccountId, validation.Required, is.UUID),
	)
}
