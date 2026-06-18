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

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
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

	groupsResp := make([]models.GroupResponse, 0)
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

func (h *Handler) GetGroupById(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

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

	group, err := h.store.JoinGroupByInviteCode(r.Context(), userId, model.InviteCode)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, group)
}

func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

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

	users := make([]models.UserResponse, 0)
	for _, v := range *userRows {
		users = append(users, models.UserResponse{
			Id:        v.Id,
			Username:  v.Username,
			Email:     v.Email,
			CreatedAt: v.CreatedAt,
		})
	}

	helpers.WriteJson(w, http.StatusOK, users)
}

func (h *Handler) RegenInviteCode(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	code, err := helpers.GenerateNewInviteCode(helpers.InviteCodeLength)
	if err != nil {
		log.Println("error: failed to generate code\n\t", err.Error())
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	group, err := h.store.UpdateGroupInviteCode(r.Context(), groupId, code)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, group)
}

func (h *Handler) PatchGroup(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

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

	group, err := h.store.UpdateGroupInfo(r.Context(), groupId, model.Name, model.Description)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, group)
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	memberIdString := chi.URLParam(r, helpers.ParamUserId)

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

	groupMember, err := h.store.UpdateMemberRole(r.Context(), groupId, memberId, models.GroupRole(model.Role))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, groupMember)
}

func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	memberIdString := chi.URLParam(r, helpers.ParamUserId)

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
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "member removed from group"})
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

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
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "group removed"})
}
