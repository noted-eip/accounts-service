package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateAddGroupNote(in *accountsv1.AddGroupNoteRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.NoteId, validation.Required, is.UUID),
		validation.Field(&in.Title, validation.Required, is.ASCII),
	)
}

func ValidateRemoveGroupNote(in *accountsv1.RemoveGroupNoteRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.NoteId, validation.Required, is.UUID),
	)
}

func ValidateUpdateGroupNote(in *accountsv1.UpdateGroupNoteRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
	)
}

func ValidateListGroupNote(in *accountsv1.ListGroupNotesRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.AuthorAccountId, validation.Required, is.UUID),
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

func ValidateGetGroupNote(in *accountsv1.GetGroupNoteRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.GroupId, validation.Required, is.UUID),
		validation.Field(&in.NoteId, validation.Required, is.UUID),
	)
}
