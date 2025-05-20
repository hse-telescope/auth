package auth

import (
	cryptorand "crypto/rand"
	"crypto/rsa"
	"math/big"
	mathrand "math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtPrivateKey  *rsa.PrivateKey
	accessTokenTTL = 15 * time.Minute
)

func InitJWT(privateKey string) error {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return err
	}
	jwtPrivateKey = key
	return nil
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jwtPrivateKey)
}

func GenerateRefreshToken(userID int64) (string, error) {
	claims := &Claims{
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jwtPrivateKey)
}

func GenerateNewPassword() (string, error) {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		symbols   = "!@#$%^&*()-_=+,.?/:;{}[]~"
	)

	length := 32
	allChars := lowercase + uppercase + digits + symbols
	password := make([]byte, length)

	password[0] = lowercase[randInt(len(lowercase))]
	password[1] = uppercase[randInt(len(uppercase))]
	password[2] = digits[randInt(len(digits))]
	password[3] = symbols[randInt(len(symbols))]

	for i := 4; i < length; i++ {
		password[i] = allChars[randInt(len(allChars))]
	}

	mathrand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password), nil
}

func randInt(max int) int {
	n, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
