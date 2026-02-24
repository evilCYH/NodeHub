package op

import (
	"context"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/auth"
	"golang.org/x/crypto/bcrypt"
)

var authRepo interfaces.AuthRepository
var authData *auth.Data

func AuthRepo() interfaces.AuthRepository {
	if authRepo == nil {
		authRepo = repo.Auth()
	}
	return authRepo
}

func AuthGet() (auth.Data, error) {
	var err error
	if authData == nil {
		authData, err = AuthRepo().Get(context.Background())
	}
	return *authData, err
}
func AuthUpdateName(name string) error {
	if authData == nil {
		AuthGet()
	}
	authData.UserName = name
	err := AuthRepo().UpdateName(context.Background(), name)
	if err != nil {
		return err
	}
	return nil
}
func AuthUpdatePassWord(password string) error {
	if authData == nil {
		AuthGet()
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	authData.Password = string(hashedBytes)
	err = AuthRepo().UpdatePassword(context.Background(), authData.Password)
	if err != nil {
		return err
	}
	return nil
}
func AuthVerify(username, password string) error {
	if authData == nil {
		AuthGet()
	}
	if authData.UserName != username {
		return fmt.Errorf("用户名不匹配")
	}
	return bcrypt.CompareHashAndPassword([]byte(authData.Password), []byte(password))
}
