package db

import (
	"context"
	"database/sql"

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
