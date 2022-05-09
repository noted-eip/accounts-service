package validator

import (
	"accounts-service/grpc/accountspb"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

//Basic check for new account information
func ValidateAccountCreation(newAccount *accountspb.CreateAccountRequest) error {
	return validation.Errors{
		"name":     validation.Validate(newAccount.Name, validation.Required, validation.Length(4, 20)),
		"email":    validation.Validate(newAccount.Email, validation.Required, is.Email),
		"password": validation.Validate(newAccount.Password, validation.Required, validation.Length(4, 20)),
	}.Filter()
}

// Compare id format from request with MongoId
func ValidateAccountDeletion(Account *accountspb.DeleteAccountRequest) error {
	return validation.Errors{
		"id": validation.Validate(Account.Id, validation.Required, is.MongoID),
	}.Filter()
}
