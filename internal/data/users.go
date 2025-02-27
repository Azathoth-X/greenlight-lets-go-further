package data

import (
	"errors"
	"greenlight/internal/validator"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hashedPassword
	return nil
}

func (p *password) Match(plaintextPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must provide an email")
	v.Check(validator.Matches(email, validator.EmailRx), "email", "must be a valid email")
}

func ValidatePlaintextPassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must provide a password ")
	v.Check(len([]byte(password)) >= 8, "password", "password must be at least 8 bytes")
	v.Check(len([]byte(password)) <= 72, "password", "password must be at most 72 bytes")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must provide a name ")
	v.Check(len([]byte(user.Name)) <= 500, "name", "must be less at most 500 bytes")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePlaintextPassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing hashed password for user")
	}
}
