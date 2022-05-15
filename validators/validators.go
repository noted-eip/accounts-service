package validator

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
	return validation.ValidateStruct(
		validation.Field(&in.Id, validation.When(in.Id != "", validation.Required), is.UUID),
		validation.Field(&in.Email, validation.When(in.Id != "", validation.Required), is.Email),
	)
}
