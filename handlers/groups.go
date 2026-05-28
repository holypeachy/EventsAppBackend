package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (h *Handler) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var model models.CreateGroupModel

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

	group, err := h.store.CreateGroup(r.Context(), userId, model.Name, model.Description)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusCreated, group)
}

func (h *Handler) GetGroupsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groups, err := h.store.GetGroupsUserBelongsTo(r.Context(), userId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	var groupsResp []models.GroupResponse
	for _, v := range *groups {
		group := models.GroupResponse{
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
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	groupResp := models.GroupResponse{
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

	var model models.JoinGroupModel
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
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
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
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	var users []models.UserResponse
	for _, v := range *userRows {
		users = append(users, models.UserResponse{
			Id:        v.Id,
			Username:  v.Username,
			Email:     v.Email,
			CreatedAt: v.CreatedAt,
		})
	}

	log.Println("log: group members sent")
	helpers.WriteJson(w, http.StatusOK, users)
}

func (h *Handler) RegenInviteCodeHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	code, err := helpers.GenerateNewInviteCode(8)
	if err != nil {
		log.Println("error: failed to generate code\n", err)
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = h.store.UpdateGroupInviteCode(r.Context(), groupId, code)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	log.Println("log: code regenerated succesfully")
	helpers.WriteJson(w, http.StatusOK, map[string]string{"inviteCode": code})
}

func (h *Handler) PatchGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var model models.PatchGroupModel
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.store.UpdateGroupInfo(r.Context(), groupId, model.Name, model.Description)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "group info updated"})
}

func (h *Handler) UpdateMemberRoleHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	memberIdString := chi.URLParam(r, "userId")

	memberId, err := uuid.Parse(memberIdString)
	if err != nil {
		log.Println("error: invalid group id", memberId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var model models.UpdateMemberRoleModel
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.store.UpdateMemberRole(r.Context(), groupId, memberId, models.GroupRole(model.Role))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "member role updated"})
}

func (h *Handler) RemoveMemberFromGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	memberIdString := chi.URLParam(r, "userId")

	memberId, err := uuid.Parse(memberIdString)
	if err != nil {
		log.Println("error: invalid group id", memberId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	err = h.store.RemoveMemberFromGroup(r.Context(), groupId, memberId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "member removed from group"})
}
func (h *Handler) DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	err = h.store.DeleteGroup(r.Context(), groupId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "group removed"})
}
