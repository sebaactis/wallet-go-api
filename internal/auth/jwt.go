package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret     []byte
	ttlNormal  time.Duration
	ttlRefresh time.Duration
}

func NewJWT() *JWT {
	sec := os.Getenv("JWT_SECRET")

	if sec == "" {
		sec = "dev-secret"
	}

	ttlMin := 60
	ttlMinRefresh := 1440

	if s := os.Getenv("JWT_TTL_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMin = n
		}
	}

	return &JWT{secret: []byte(sec), ttlNormal: time.Duration(ttlMin) * time.Minute, ttlRefresh: time.Duration(ttlMinRefresh) * time.Minute}
}

func (j *JWT) Sign(userID uint, email string, tokenType TokenType) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   j.getExpiration(tokenType, now),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString(j.secret)
}

func (j *JWT) Parse(token string) (uint, string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("bad signing method")
		}

		return j.secret, nil
	})

	if err != nil || !parsed.Valid {
		return 0, "", errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)

	if !ok {
		return 0, "", errors.New("invalid claims")
	}

	subF, ok := claims["sub"].(float64)

	if !ok {
		return 0, "", errors.New("missing sub")
	}

	email, _ := claims["email"].(string)

	return uint(subF), email, nil
}

func (j *JWT) getExpiration(tokenType TokenType, now time.Time) int64 {
	if tokenType == TokenTypeAccess {
		return now.Add(j.ttlNormal).Unix()
	}
	return now.Add(j.ttlRefresh).Unix()
}
