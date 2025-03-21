package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"greenlight/internal/validator"
	"time"
)

const (
	ScopeAuthentication = "authentication"
	ScopeActivation     = "activation"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randBytes := make([]byte, 16)

	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "token", "must be provided")
	v.Check(len(plaintext) == 26, "token", "length of token must be 26")
}

type TokenModel struct {
	DB *sql.DB
}

func (m *TokenModel) Insert(token *Token) error {

	stmt := `INSERT INTO tokens (hash,user_id,expiry,scope)
			 VALUES ($1,$2,$3,$4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, args...)
	return err
}

func (m *TokenModel) DeleteAllForUser(scope string, userID int64) error {

	stmt := `DELETE FROM tokens
			 WHERE user_id=$1 AND scope=$2`

	args := []any{userID, scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, args...)
	return err
}

func (m *TokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
