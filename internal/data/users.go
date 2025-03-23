package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"greenlight/internal/validator"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

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

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {

	stmt := `	INSERT INTO users (name,email,password_hash,activated)
				VALUES ($1,$2,$3,$4)
				RETURNING id,created_at,version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail

		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	stmt := `
			SELECT id,created_at,name,email,password_hash,activated,version
			FROM users
			WHERE email=$1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}

func (m UserModel) Update(user *User) error {

	stmt := `
        UPDATE users 
        SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil

}

func (m *UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {

	smtm := `SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
			FROM users
			INNER JOIN tokens
			ON users.id = tokens.user_id
			WHERE tokens.hash = $1
			AND tokens.scope = $2 
			AND tokens.expiry > $3`

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	args := []any{tokenHash[:], tokenScope, time.Now()}

	user := &User{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, smtm, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return user, nil

}
