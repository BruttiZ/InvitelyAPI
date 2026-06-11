package jwt

import "time"

type Claims struct {
	Subject   string
	ExpiresAt time.Time
}

func Sign(claims Claims, secret string) (string, error) {
	return "", nil
}

func Verify(token, secret string) (Claims, error) {
	return Claims{}, nil
}
