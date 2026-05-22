package handlers

import (
	"errors"
	"strings"
)

func (model *RegisterModel) Validate() error {
	if strings.TrimSpace(model.Username) == "" {
		return errors.New("username cannot be empty")
	}
	if strings.TrimSpace(model.Email) == "" {
		return errors.New("email cannot be empty")
	}
	if strings.TrimSpace(model.Password) == "" {
		return errors.New("password cannot be empty")
	}

	return nil
}

func (model *LoginModel) Validate() error {
	if strings.TrimSpace(model.Email) == "" {
		return errors.New("email cannot be empty")
	}
	if strings.TrimSpace(model.Password) == "" {
		return errors.New("password cannot be empty")
	}

	return nil
}
