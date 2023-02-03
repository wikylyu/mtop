package db

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
 * return nil if user not found.
 */
func GetUserByUsername(username string) (*User, error) {
	user := User{Username: username}

	if err := DB().NewSelect().Model(&user).WherePK().Scan(context.Background()); err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("DB Error: %v", err)
			return nil, err
		}
		return nil, nil
	}

	return &user, nil
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randomSalt(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func CreateUser(username, password, salt string, update bool) (*User, error) {
	if salt == "" {
		salt = randomSalt(12)
	}
	user := User{
		Username: username,
		Salt:     salt,
	}
	user.Password = user.encryptPassword(password)
	q := DB().NewInsert().Model(&user)

	if update {
		q = q.On(`CONFLICT (username) DO UPDATE`).
			Set(`"salt" = ?`, user.Salt).Set(`"password"=?`, user.Password)
	}

	if _, err := q.Exec(context.Background()); err != nil {
		log.Errorf("DB Error: %v", err)
		return nil, err
	}
	return &user, nil
}

func DeleteUser(username string) error {
	user := User{
		Username: username,
	}
	if _, err := DB().NewDelete().Model(&user).WherePK().Exec(context.Background()); err != nil {
		log.Errorf("DB Error: %v", err)
		return err
	}
	return nil
}
