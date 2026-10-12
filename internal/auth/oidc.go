package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidToken   = errors.New("invalid or malformed token")
	ErrTokenExpired   = errors.New("token has expired")
	ErrInvalidIssuer  = errors.New("invalid token issuer")
	ErrInvalidAudience = errors.New("invalid token audience")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrUnauthorized   = errors.New("unauthorized")
)

// Claims represents standard JWT claims and Corridor RBAC attributes.
type Claims struct {
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Role      Role   `json:"role,omitempty"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	Agency    string `json:"agency,omitempty"`
}

// Authenticator validates bearer tokens and API keys.
type Authenticator interface {
	AuthenticateToken(ctx context.Context, tokenString string) (*User, error)
	AuthenticateAPIKey(ctx context.Context, apiKey string) (*User, error)
	AuthenticateRequest(r *http.Request) (*User, error)
}

// Config configures token and API key validation.
type Config struct {
	Issuer   string
	Audience string
	Secret   string            // HMAC secret for signature verification
	APIKeys  map[string]User   // Static API keys mapped to User identities
}

type tokenAuthenticator struct {
	cfg Config
}

// NewAuthenticator creates a new Authenticator from Config.
func NewAuthenticator(cfg Config) Authenticator {
	if cfg.APIKeys == nil {
		cfg.APIKeys = make(map[string]User)
	}
	return &tokenAuthenticator{cfg: cfg}
}

// AuthenticateToken verifies a JWT token string.
func (a *tokenAuthenticator) AuthenticateToken(_ context.Context, tokenString string) (*User, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, ErrInvalidToken
	}

	if header.Alg != "HS256" {
		return nil, fmt.Errorf("%w: unsupported algorithm %s", ErrInvalidToken, header.Alg)
	}

	// Verify signature
	if a.cfg.Secret != "" {
		signedData := parts[0] + "." + parts[1]
		mac := hmac.New(sha256.New, []byte(a.cfg.Secret))
		mac.Write([]byte(signedData))
		expectedSig := mac.Sum(nil)

		actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil || subtle.ConstantTimeCompare(actualSig, expectedSig) != 1 {
			return nil, ErrInvalidSignature
		}
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	now := time.Now().Unix()
	if claims.ExpiresAt > 0 && claims.ExpiresAt < now {
		return nil, ErrTokenExpired
	}
	if a.cfg.Issuer != "" && claims.Issuer != "" && claims.Issuer != a.cfg.Issuer {
		return nil, ErrInvalidIssuer
	}
	if a.cfg.Audience != "" && claims.Audience != "" && claims.Audience != a.cfg.Audience {
		return nil, ErrInvalidAudience
	}

	role := claims.Role
	if role == "" {
		role = RolePublic
	}

	return &User{
		Subject: claims.Subject,
		Role:    role,
		Email:   claims.Email,
		Name:    claims.Name,
		Agency:  claims.Agency,
	}, nil
}

// AuthenticateAPIKey verifies a raw API key against configured keys.
func (a *tokenAuthenticator) AuthenticateAPIKey(_ context.Context, apiKey string) (*User, error) {
	if apiKey == "" {
		return nil, ErrUnauthorized
	}
	user, ok := a.cfg.APIKeys[apiKey]
	if !ok {
		return nil, ErrUnauthorized
	}
	return &user, nil
}

// AuthenticateRequest extracts credentials from Authorization or X-API-Key headers.
func (a *tokenAuthenticator) AuthenticateRequest(r *http.Request) (*User, error) {
	// 1. Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return a.AuthenticateAPIKey(r.Context(), apiKey)
	}

	// 2. Check Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return a.AuthenticateToken(r.Context(), parts[1])
		}
	}

	// Unauthenticated defaults to nil
	return nil, nil
}

// SignToken creates a signed JWT string using HMAC-SHA256 for testing and token distribution.
func SignToken(claims Claims, secret string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	part1 := base64.RawURLEncoding.EncodeToString(headerJSON)
	part2 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signedData := part1 + "." + part2

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedData))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signedData + "." + signature, nil
}
