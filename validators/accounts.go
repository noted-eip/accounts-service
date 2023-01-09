package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateCreateAccountRequest(in *accountsv1.CreateAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Name, validation.Required, validation.Length(4, 20)),
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateGetAccountRequest(in *accountsv1.GetAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Id, validation.When(in.Email == "", validation.Required), is.UUID),
		validation.Field(&in.Email, validation.When(in.Id == "", validation.Required), is.Email),
	)
}

func ValidateUpdateAccountRequest(in *accountsv1.UpdateAccountRequest) error {
	return validation.Validate(in.Account.Id,
		validation.Required,
		is.UUID,
	)
}

func ValidateDeleteAccountRequest(in *accountsv1.DeleteAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Id, validation.Required, is.UUID),
	)
}

func ValidateAuthenticateRequest(in *accountsv1.AuthenticateRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateListRequest(in *accountsv1.ListAccountsRequest) error {
	err := validation.Validate(in.Limit, validation.When(in.Limit != 0, validation.Required), validation.Min(0))
	if err != nil {
		return err
	}
	err = validation.Validate(in.Offset, validation.When(in.Offset != 0, validation.Required), validation.Min(0))
	if err != nil {
		return err
	}

	err = validation.Validate(in.EmailContains, validation.Required)
	if err != nil {
		return err
	}
	return nil
}
