package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"golang.org/x/crypto/bcrypt"
)

type RegisterModel struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
	User         *LoginResponseUser `json:"user"`
}

type LoginResponseUser struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var model RegisterModel

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

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(model.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}
	user, err := h.store.RegisterUser(r.Context(), model.Username, model.Email, string(hashedPass))
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	accessToken, rawRefresh, err := h.issueTokens(r.Context(), user.Id)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUser := LoginResponseUser{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         &respUser,
	}

	log.Println("log: user registered,", user.Username)
	helpers.WriteJson(w, http.StatusCreated, resp)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var model LoginModel

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
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(model.Password))
	if err != nil {
		log.Println("error: password is incorrect,", user.Username)
		helpers.WriteErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, rawRefresh, err := h.issueTokens(r.Context(), user.Id)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respUser := LoginResponseUser{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         &respUser,
	}

	log.Println("log: user logged in,", user.Username)
	helpers.WriteJson(w, http.StatusOK, resp)
}

func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (h *Handler) issueTokens(ctx context.Context, userId uuid.UUID) (accessToken string, rawRefresh string, err error) {
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
