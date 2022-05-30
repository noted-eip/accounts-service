package main

// func TestAccountsServiceCreateAccount(t *testing.T) {
// 	srv := accountsService{
// 		auth:   auth.NewService(genKeyOrFail(t)),
// 		logger: zap.NewNop().Sugar(),
// 	}

// 	res, err := srv.CreateAccount(context.TODO(), &accountspb.CreateAccountRequest{Email: "mail.test@gmail.com", Password: "password", Name: "Maxime"})
// 	require.Nil(t, err)
// 	require.Empty(t, res)
// }

// func genKeyOrFail(t *testing.T) ed25519.PrivateKey {
// 	_, priv, err := ed25519.GenerateKey(nil)
// 	require.NoError(t, err)
// 	return priv
// }
