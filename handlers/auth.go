package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var model models.RegisterModel

	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(model.Password), auth.BcryptCost)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}
	user, err := h.store.RegisterUser(r.Context(), model.Username, model.Email, string(hashedPass))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	accessToken, rawRefresh, err := h.issueLoginTokens(r.Context(), user.Id)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUser := models.LoginUserResponse{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}

	resp := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         &respUser,
	}

	log.Println("log: user registered with username", user.Username)
	helpers.WriteJson(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var model models.LoginModel

	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), model.Email)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(model.Password))
	if err != nil {
		log.Println("error: password is incorrect for user", user.Username)
		helpers.WriteErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, rawRefresh, err := h.issueLoginTokens(r.Context(), user.Id)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUser := models.LoginUserResponse{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}

	resp := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         &respUser,
	}

	log.Println("log: user logged in,", user.Username)
	helpers.WriteJson(w, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var model models.RefreshModel

	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.store.GetRefreshRowByHash(r.Context(), auth.HashRefreshToken(model.RefreshToken))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	if time.Now().After(token.ExpiresAt) {
		log.Println("log: token expired, login again")
		err := h.store.DeleteRefreshTokenById(r.Context(), token.Id)
		if err != nil {
			helpers.HandlePgxError(err)
		}
		helpers.WriteErr(w, http.StatusUnauthorized, "must login again")
		return
	}

	access, err := auth.CreateAccessToken(token.UserId, h.jwtSecret)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}
	log.Println("log: refresh token valid, sending new access token")
	helpers.WriteJson(w, http.StatusOK, map[string]string{"accessToken": access})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var model models.LogoutModel

	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.store.DeleteRefreshTokenByHash(r.Context(), auth.HashRefreshToken(model.RefreshToken))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	log.Println("log: user logged out")
	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "user successfully logged out"})
}

func (h *Handler) issueLoginTokens(ctx context.Context, userId uuid.UUID) (accessToken string, rawRefresh string, err error) {
	rawRefresh, err = auth.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	hashedRefresh := auth.HashRefreshToken(rawRefresh)

	err = h.store.StoreRefreshToken(ctx, userId, hashedRefresh)
	if err != nil {
		return "", "", err
	}

	accessToken, err = auth.CreateAccessToken(userId, h.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, rawRefresh, nil
}
