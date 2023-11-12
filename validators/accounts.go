package validators

import (
	accountsv1 "accounts-service/protorepo/noted/accounts/v1"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type notSameRecipientAndSenderRule struct{}

func ValidateCreateAccountRequest(in *accountsv1.CreateAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Name, validation.Required, validation.Length(4, 20)),
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateGetAccountRequest(in *accountsv1.GetAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.AccountId, validation.When(in.Email == "", validation.Required)),
		validation.Field(&in.Email, validation.When(in.AccountId == "", validation.Required), is.Email),
	)
}

func ValidateGetMailsFromIDs(in *accountsv1.GetMailsFromIDsRequest) error {
	var err error = nil

	if len(in.AccountsIds) <= 0 {
		return errors.New("Accounts Id list is nil or empty")
	}
	for _, id := range in.AccountsIds {
		err = validation.Validate(id, validation.Required)
	}
	return err
}

func ValidateUpdateAccountRequest(in *accountsv1.UpdateAccountRequest) error {
	err := validation.Validate(in.AccountId, validation.Required)
	if err != nil {
		return err
	}
	err = validation.Validate(in.Account, validation.Required)
	if err != nil {
		return err
	}
	return validation.Validate(in.Account.Name, validation.Required, validation.Length(4, 20))
}

func ValidateDeleteAccountRequest(in *accountsv1.DeleteAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.AccountId, validation.Required),
	)
}

func ValidateRegisterUserToMobileBeta(in *accountsv1.RegisterUserToMobileBetaRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.AccountId, validation.Required),
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
	return nil
}

func ValidateForgetAccountPasswordRequest(in *accountsv1.ForgetAccountPasswordRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email))
}

func ValidateForgetAccountPasswordValidateTokenRequest(in *accountsv1.ForgetAccountPasswordValidateTokenRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Token, validation.Required, validation.Length(4, 4)),
		validation.Field(&in.AccountId, validation.Required, validation.NotNil))
}

func ValidateUpdateAccountPasswordRequest(in *accountsv1.UpdateAccountPasswordRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.AccountId, validation.Required),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
		validation.Field(&in.Token, validation.When(in.Token != ""), validation.Length(4, 4)),
		validation.Field(&in.OldPassword, validation.When(in.OldPassword != ""), validation.Length(4, 20)),
	)
}

func ValidateSendGroupInviteMail(in *accountsv1.SendGroupInviteMailRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.RecipientId, validation.Required, validation.NotNil),
		validation.Field(&in.SenderId, validation.Required, validation.NotNil),
		validation.Field(&in.GroupName, validation.Required, validation.NotNil),
	)
	if err != nil {
		return err
	}

	if in.RecipientId == in.SenderId {
		return errors.New("recipient and sender IDs cannot be the same")
	}
	return nil
}

func ValidateAuthenticateGoogleRequest(in *accountsv1.AuthenticateGoogleRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.ClientAccessToken, validation.Required, validation.NotNil),
	)
}

func ValidateAccountValidationStateRequest(in *accountsv1.ValidateAccountRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateSendValidationToken(in *accountsv1.SendValidationTokenRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}

func ValidateIsAccountValidateRequest(in *accountsv1.IsAccountValidateRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.Email, validation.Required, is.Email),
		validation.Field(&in.Password, validation.Required, validation.Length(4, 20)),
	)
}
