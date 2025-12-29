package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Secret []byte
	Issuer string

	AccessTTL  time.Duration // напр. 15 * time.Minute
	RefreshTTL time.Duration // напр. 30 * 24 * time.Hour
	Audience   string        // напр. "autera"
}

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

type Claims struct {
	UserID       int64     `json:"user_id"`
	Roles        []string  `json:"roles"`
	TokenVersion int64     `json:"token_version"`
	TokenType    TokenType `json:"token_type"`
	DeviceID     string    `json:"device_id,omitempty"`

	jwt.RegisteredClaims
}

type JWT struct{ cfg JWTConfig }

func NewJWT(cfg JWTConfig) *JWT { return &JWT{cfg: cfg} }

func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (j *JWT) IssueAccess(userID int64, roles []string, tokenVersion int64, deviceID string) (string, string, time.Time, error) {
	jti, err := newJTI()
	if err != nil {
		return "", "", time.Time{}, err
	}

	now := time.Now()
	exp := now.Add(j.cfg.AccessTTL)

	claims := Claims{
		UserID:       userID,
		Roles:        roles,
		TokenVersion: tokenVersion,
		TokenType:    TokenAccess,
		DeviceID:     deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Audience:  jwt.ClaimStrings{j.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(j.cfg.Secret)
	return s, jti, exp, err
}

func (j *JWT) IssueRefresh(userID int64, tokenVersion int64, deviceID string) (string, string, time.Time, error) {
	jti, err := newJTI()
	if err != nil {
		return "", "", time.Time{}, err
	}

	now := time.Now()
	exp := now.Add(j.cfg.RefreshTTL)

	claims := Claims{
		UserID:       userID,
		TokenVersion: tokenVersion,
		TokenType:    TokenRefresh,
		DeviceID:     deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Audience:  jwt.ClaimStrings{j.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(j.cfg.Secret)
	return s, jti, exp, err
}

func (j *JWT) Parse(tokenStr string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(j.cfg.Issuer),
		jwt.WithAudience(j.cfg.Audience),
	)

	claims := &Claims{}
	_, err := parser.ParseWithClaims(tokenStr, claims, func(_ *jwt.Token) (any, error) {
		return j.cfg.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenAccess && claims.TokenType != TokenRefresh {
		return nil, errors.New("invalid token_type")
	}
	if claims.ID == "" {
		return nil, errors.New("missing jti")
	}
	return claims, nil
}
