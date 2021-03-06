package validators

import (
	"accounts-service/grpc/accountspb"

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
		validation.Field(&in.Email, validation.When(in.Id == "", validation.Required), is.Email),
	)
}

func ValidateUpdateAccountRequest(in *accountspb.UpdateAccountRequest) error {
	return validation.Validate(in.Account.Id,
		validation.Required,
		is.UUID,
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
