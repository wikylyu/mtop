package db

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:mtop_user"`

	Username string `bun:"username,notnull,pk" json:"username"`
	Salt     string `bun:"salt,notnull" json:"-"`
	Password string `bun:"password,notnull" json:"password"`

	CreatedTime time.Time `bun:"created_time,notnull,default:current_timestamp" json:"created_time"`
	UpdatedTime time.Time `bun:"updated_time,notnull,default:current_timestamp" json:"updated_time"`
}

func (u *User) Auth(plain string) bool {
	if u.Password == "" {
		return false
	}
	h := sha256.New()
	h.Write([]byte(plain))
	return u.Password == fmt.Sprintf("%x", h.Sum(nil))
}
