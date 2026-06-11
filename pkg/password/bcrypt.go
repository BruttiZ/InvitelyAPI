package password

import "errors"

var ErrPasswordNotConfigured = errors.New("password hashing is not configured")

func Hash(plain string) (string, error) {
	return "", ErrPasswordNotConfigured
}

func Compare(hash, plain string) error {
	return ErrPasswordNotConfigured
}
