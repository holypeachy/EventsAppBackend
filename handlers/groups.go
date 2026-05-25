package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
)

type CreateGroupModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GroupResponse struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	InviteCode  string    `json:"inviteCode"`
}

type JoinGroupModel struct {
	InviteCode string `json:"inviteCode"`
}

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *Handler) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var model CreateGroupModel

	err = json.NewDecoder(r.Body).Decode(&model)
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

	groupId, err := h.store.CreateGroup(r.Context(), userId, model.Name, model.Description)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("log: group created, ", groupId)
	helpers.WriteJson(w, http.StatusOK, map[string]string{"groupId": groupId.String()})
}

func (h *Handler) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groups, err := h.store.GetGroupsUserBelongs(r.Context(), userId)
	if err != nil {
		log.Println("error: failed to get grops user belongs to")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var groupsResp []GroupResponse
	for _, v := range *groups {
		group := GroupResponse{
			Id:          v.Id,
			Name:        v.Name,
			Description: v.Description,
			CreatedBy:   v.CreatedBy,
			CreatedAt:   v.CreatedAt,
			InviteCode:  v.InviteCode,
		}

		groupsResp = append(groupsResp, group)
	}

	helpers.WriteJson(w, http.StatusOK, groupsResp)
}

func (h *Handler) GetGroupByIdHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	groupRow, err := h.store.GetGroupById(r.Context(), groupId)
	if err != nil {
		log.Println("error: failed to get group by id\n", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groupResp := GroupResponse{
		Id:          groupRow.Id,
		Name:        groupRow.Name,
		Description: groupRow.Description,
		CreatedBy:   groupRow.CreatedBy,
		CreatedAt:   groupRow.CreatedAt,
		InviteCode:  groupRow.InviteCode,
	}

	log.Println("log: group by id success")
	helpers.WriteJson(w, http.StatusOK, groupResp)
}

func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var model JoinGroupModel
	err = json.NewDecoder(r.Body).Decode(&model)
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

	err = h.store.JoinGroupByInviteCode(r.Context(), userId, model.InviteCode)
	if err != nil {
		log.Println("error: failed to join by invite", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "user joined group"})
}

func (h *Handler) GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	userRows, err := h.store.GetGroupMembers(r.Context(), groupId)
	if err != nil {
		log.Println("error: failed to get group members\n", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var users []UserResponse
	for _, v := range *userRows {
		users = append(users, UserResponse{
			Id:        v.Id,
			Username:  v.Username,
			Email:     v.Email,
			CreatedAt: v.CreatedAt,
		})
	}

	log.Println("log: group members sent")
	helpers.WriteJson(w, http.StatusOK, users)
}
