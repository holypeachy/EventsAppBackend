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

func (model *RefreshModel) Validate() error {
	if strings.TrimSpace(model.RefreshToken) == "" {
		return errors.New("refreshToken cannot be empty")
	}

	return nil
}

func (model *LogoutModel) Validate() error {
	if strings.TrimSpace(model.RefreshToken) == "" {
		return errors.New("refreshToken cannot be empty")
	}

	return nil
}

func (model *CreateGroupModel) Validate() error {
	if strings.TrimSpace(model.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if strings.TrimSpace(model.Description) == "" {
		return errors.New("description cannot be empty")
	}

	return nil
}

func (model *JoinGroupModel) Validate() error {
	if strings.TrimSpace(model.InviteCode) == "" {
		return errors.New("invite code cannot be empty")
	}

	return nil
}
