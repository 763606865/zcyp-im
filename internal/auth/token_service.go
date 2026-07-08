package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"zcyp-im/internal/config"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	AppCode string `json:"app_code"`
	UserID  string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret      []byte
	issuer      string
	expireHours int
}

func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{
		secret:      []byte(cfg.Secret),
		issuer:      cfg.Issuer,
		expireHours: cfg.ExpireHours,
	}
}

func (s *TokenService) Issue(appCode, userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.expireHours) * time.Hour)
	claims := Claims{
		AppCode: appCode,
		UserID:  userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (s *TokenService) Parse(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}
