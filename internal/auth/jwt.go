package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"` // Agregar este campo
	jwt.RegisteredClaims
}

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

	ttlMin := 1
	ttlMinRefresh := 1440

	if s := os.Getenv("JWT_TTL_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMin = n
		}
	}

	if s := os.Getenv("JWT_TTL_REFRESH_MINUTES"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			ttlMinRefresh = n
		}
	}

	return &JWT{secret: []byte(sec), ttlNormal: time.Duration(ttlMin) * time.Minute, ttlRefresh: time.Duration(ttlMinRefresh) * time.Minute}
}

func (j *JWT) Sign(userID uint, email string, tokenType TokenType) (string, error) {
	now := time.Now()

	claims := Claims{
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(j.getExpiration(tokenType, now)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWT) Parse(tokenIn string) (uint, string, TokenType, error) {
	token, err := jwt.ParseWithClaims(
		tokenIn,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return j.secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil || !token.Valid {
		return 0, "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, "", "", errors.New("invalid token claims")
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return 0, "", "", errors.New("token expired")
	}

	idU64, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, "", "", errors.New("invalid subject")
	}

	return uint(idU64), claims.Email, claims.TokenType, nil
}

func (j *JWT) getExpiration(tokenType TokenType, now time.Time) time.Time {
	if tokenType == TokenTypeAccess {
		return now.Add(j.ttlNormal)
	}
	return now.Add(j.ttlRefresh)
}

func (j *JWT) GetTTL(tokenType TokenType) time.Duration {
	var ttl time.Duration

	if tokenType == TokenTypeAccess {
		ttl = j.ttlNormal
	} else {
		ttl = j.ttlRefresh
	}

	return ttl
}
