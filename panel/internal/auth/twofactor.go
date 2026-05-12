package auth

import "github.com/pquerna/otp/totp"

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

func GenerateTOTPSecret(accountName string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Pelican",
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}
