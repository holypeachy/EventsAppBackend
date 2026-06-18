package main

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/holypeachy/EventsAppBackend/handlers"
	mid "github.com/holypeachy/EventsAppBackend/middleware"
)

func registerRoutes(r chi.Router, h *handlers.Handler, m *mid.Middleware) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.Health)

		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(50, time.Minute))

			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Post("/refresh", h.Refresh)
			r.Post("/logout", h.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(m.RequireAuth)

			r.Post("/groups", h.CreateGroup)
			r.Get("/groups", h.GetGroups)

			r.Post("/groups/join", h.JoinGroup)
			r.Get("/events", h.GetEvents)

			r.Group(func(r chi.Router) {
				r.Use(m.RequireGroupMember)

				r.Get("/groups/{groupId}", h.GetGroupById)
				r.Get("/groups/{groupId}/members", h.GetGroupMembers)
				r.Post("/groups/{groupId}/events", h.CreateEvent)
				r.Get("/groups/{groupId}/events", h.GetGroupEvents)
			})

			r.Group(func(r chi.Router) {
				r.Use(m.RequireGroupAdmin)

				r.Post("/groups/{groupId}/invite-code/regen", h.RegenInviteCode)
				r.Patch("/groups/{groupId}", h.PatchGroup)
				r.Patch("/groups/{groupId}/members/{userId}", h.UpdateMemberRole)
				r.Delete("/groups/{groupId}/members/{userId}", h.RemoveGroupMember)
			})

			r.Group(func(r chi.Router) {
				r.Use(m.RequireGroupOwner)

				r.Delete("/groups/{groupId}", h.DeleteGroup)
			})

			r.Group(func(r chi.Router) {
				r.Use(m.RequireEventParticipant)

				r.Get("/events/{eventId}", h.GetEventById)
				r.Get("/events/{eventId}/participants", h.GetEventParticipants)
				r.Patch("/events/{eventId}/participants/{userId}/rsvp", h.Rsvp)
			})

			r.Group(func(r chi.Router) {
				r.Use(m.RequireEventManager)

				r.Patch("/events/{eventId}", h.PatchEvent)
				r.Delete("/events/{eventId}", h.DeleteEvent)
				r.Delete("/events/{eventId}/participants/{userId}", h.RemoveParticipant)

				r.Post("/events/{eventId}/participants", h.AddParticipant)
			})

			r.Group(func(r chi.Router) {
				r.Use(m.RequireEventOwner)

			})
		})

	})
}
