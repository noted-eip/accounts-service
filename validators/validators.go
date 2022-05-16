package validators

import (
	"accounts-service/grpc/accountspb"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func ValidateCreateAccountRequest(in *accountspb.CreateAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Name, validation.Required, validation.Length(4, 20)),
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateGetAccountRequest(in *accountspb.GetAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Id, validation.When(in.Email == "", validation.Required), is.UUID),
		validation.Field(&in.Email, validation.By(func(value interface{}) error {
			return errors.New("get account by email is unimplemented")
		})),
	)
}

func ValidateUpdateAccountRequest(in *accountspb.UpdateAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Account.Id, validation.Required, is.UUID),
	)
}

func ValidateDeleteAccountRequest(in *accountspb.DeleteAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Id, validation.Required, is.UUID),
	)
}

func ValidateAuthenticateRequest(in *accountspb.AuthenticateRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}
