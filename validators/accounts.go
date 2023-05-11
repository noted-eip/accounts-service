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
		validation.Field(&in.AccountId, validation.When(in.Email == "", validation.Required)),
		validation.Field(&in.Email, validation.When(in.AccountId == "", validation.Required), is.Email),
	)
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

func ValidateSendGoupInviteMail(in *accountsv1.SendGroupInviteMailRequest) error {
	return validation.ValidateStruct(in,
		validation.Field(&in.RecipientId, validation.Required, validation.NotNil),
		validation.Field(&in.SenderId, validation.Required, validation.NotNil),
		validation.Field(&in.GroupName, validation.Required, validation.NotNil),
		validation.Field(&in.RecipientId != &in.SenderId, validation.Required),
		validation.Field(&in.InviteLink, validation.Required, validation.NotNil),
	)
}
