package db

import "testing"

func TestUser(t *testing.T) {
	username := "test1111"
	password := "abc123321"
	user := User{
		Username: username,
		Salt:     randomSalt(10),
	}
	user.Password = user.encryptPassword(password)

	if !user.Auth(password) {
		t.Fail()
	}
	if user.Auth(password + password) {
		t.Fail()
	}

	user.Salt = randomSalt(12)
	if user.Auth(password) {
		t.Fail()
	}
}
