# Friend Event Planner — REST API Signatures

Companion to `friend_event_planner_db_notes_trimmed.md` (data model) and
`friend_event_planner_authorization.md` (access rules). Each endpoint lists
verb, path, parameters, request body, a one-line description, and an `Auth:`
line referencing the authorization doc.

---

## Conventions

```txt
Base URL        /api/v1
Format          JSON request and response bodies (application/json)
Auth header     Authorization: Bearer <access token>
                (all endpoints except register / login / refresh)
ID format       UUID (path params {…Id} are UUIDs)
Timestamps      ISO-8601 UTC (e.g. 2026-05-18T19:30:00Z)
Pagination      list endpoints accept ?page=<n>&per_page=<n>
                (defaults: page=1, per_page=20)
Money           *_cents are integers (USD cents)
```

Status codes: `200/201` success · `204` success, no body ·
`400` malformed · `401` missing/invalid token · `403` forbidden ·
`404` hidden or missing · `409` conflict/state · `422` validation error.

`?` after a field = optional. `[]` = array.

---

## Auth

```txt
POST /api/v1/auth/register
  Body: { username, email, password }
  Creates a user account. Returns the created user (no password).
  Auth: Anonymous

POST /api/v1/auth/login
  Body: { emailOrUsername, password }
  Returns { accessToken, refreshToken, user }.
  Auth: Anonymous

POST /api/v1/auth/refresh
  Body: { refreshToken }
  Rotates the refresh token; returns a new { accessToken, refreshToken }.
  Old refresh token is revoked.
  Auth: Valid, non-revoked refresh token

POST /api/v1/auth/logout
  Body: { refreshToken }
  Revokes the supplied refresh token (sets revoked_at).
  Auth: Owner of the refresh token
```

## Users

```txt
GET /api/v1/users/me
  Returns the authenticated user's profile.
  Auth: Self

PATCH /api/v1/users/me
  Body: { username?, email?, password? }
  Updates own profile. Changing password rotates password_hash.
  Auth: Self

GET /api/v1/users
  Query: query (search string), page?, per_page?
  Directory lookup for choosing invitees. Returns limited fields
  ({ id, username }) only.
  Auth: Any authenticated user
```

## Groups

```txt
POST /api/v1/groups
  Body: { name }
  Creates a group; the creator is added as group_members.role='owner'.
  Auth: Any authenticated user

GET /api/v1/groups
  Query: page?, per_page?
  Lists groups the authenticated user belongs to.
  Auth: Any authenticated user

GET /api/v1/groups/{groupId}
  Returns one group.
  Auth: Group member

PATCH /api/v1/groups/{groupId}
  Body: { name }
  Updates group fields.
  Auth: Group owner or admin

DELETE /api/v1/groups/{groupId}
  Deletes the group (cascades to members, events, …).
  Auth: Group owner
```

## Group members

```txt
GET /api/v1/groups/{groupId}/members
  Query: page?, per_page?
  Lists members with their roles.
  Auth: Group member

POST /api/v1/groups/{groupId}/members
  Body: { userId, role? }   role defaults to 'member'
  Adds a user to the group.
  Auth: Group owner or admin

PATCH /api/v1/groups/{groupId}/members/{userId}
  Body: { role }            member | admin | owner
  Changes a member's role. Transferring 'owner' moves ownership.
  Auth: Group owner

DELETE /api/v1/groups/{groupId}/members/{userId}
  Removes a member. A non-owner may remove themselves (leave).
  The owner cannot be removed without an ownership transfer first.
  Auth: Group owner/admin, or self-leave
```

## Events

```txt
POST /api/v1/groups/{groupId}/events
  Body: { name, description?, location?, startsAt, endsAt?,
          participantIds[] }
  Creates an event. The creator gets an event_participants row with
  role='owner'; each invitee gets a row (status='invited',
  role='participant') and an event_invited notification. Excluded
  members get no row (no exclusions table).
  Auth: Group member

GET /api/v1/groups/{groupId}/events
  Query: status?, page?, per_page?
  Lists this group's events the requester participates in.
  Auth: Group member (participant-filtered)

GET /api/v1/events
  Query: status?, from?, to?, page?, per_page?
  Cross-group feed: every event the requester participates in.
  Auth: Authenticated (participant-filtered)

GET /api/v1/events/{eventId}
  Returns one event. Non-participants get 404.
  Auth: Event participant

PATCH /api/v1/events/{eventId}
  Body: { name?, description?, location?, startsAt?, endsAt? }
  Updates event fields. Emits event_updated notifications.
  Auth: Event manager

POST /api/v1/events/{eventId}/cancel
  Sets events.status='cancelled'; emits event_cancelled
  notifications to participants.
  Auth: Event manager
```

## Event participants / RSVP / manager promotion

```txt
GET /api/v1/events/{eventId}/participants
  Query: status?, page?, per_page?
  Lists participants with status and role.
  Auth: Event participant

POST /api/v1/events/{eventId}/participants
  Body: { participantIds[] }
  Invites more group members; creates event_participants rows
  (status='invited') and event_invited notifications.
  Auth: Event manager

DELETE /api/v1/events/{eventId}/participants/{userId}
  Removes a participant (uninvites). A participant may remove only
  themselves (leave); the event creator cannot be removed.
  Auth: Event manager, or self-leave

PATCH /api/v1/events/{eventId}/participants/{userId}/rsvp
  Body: { status }          going | maybe | declined
  Sets the caller's RSVP; sets responded_at. {userId} must equal
  the authenticated user.
  Auth: Self only (that participant)

PATCH /api/v1/events/{eventId}/participants/{userId}/role
  Body: { role }            participant | manager
  Promotes a participant to event manager or demotes back.
  Auth: Event manager
```

## Contributions

```txt
POST /api/v1/events/{eventId}/contributions
  Body: { name, description?, type, targetAmountCents?,
          targetQuantity? }
  type ∈ money|food|item|other. targetAmountCents for money;
  targetQuantity for food/item/other. status starts 'open'.
  Auth: Event manager

GET /api/v1/events/{eventId}/contributions
  Query: status?, type?, page?, per_page?
  Lists contributions with derived progress (sum of claims).
  Auth: Event participant

GET /api/v1/contributions/{contributionId}
  Returns one contribution with progress.
  Auth: Participant of the contribution's event

PATCH /api/v1/contributions/{contributionId}
  Body: { name?, description?, type?, targetAmountCents?,
          targetQuantity?, status? }
  Updates a contribution (incl. cancel via status='cancelled').
  Auth: Event manager

DELETE /api/v1/contributions/{contributionId}
  Deletes a contribution (cascades to its claims).
  Auth: Event manager
```

## Contribution claims

```txt
POST /api/v1/contributions/{contributionId}/claims
  Body: { amountCents?, quantity?, note? }
  At least one of amountCents / quantity required (matching the
  contribution type). Multiple claims per user allowed (pools).
  Emits contribution_claimed; may flip status open→claimed→fulfilled.
  Auth: Invited participant of the contribution's event

GET /api/v1/contributions/{contributionId}/claims
  Query: page?, per_page?
  Lists all claims for the contribution.
  Auth: Participant of the contribution's event

PATCH /api/v1/contributions/{contributionId}/claims/{claimId}
  Body: { amountCents?, quantity?, note? }
  Updates a claim; recomputes contribution progress/status.
  Auth: Claim owner or event manager

DELETE /api/v1/contributions/{contributionId}/claims/{claimId}
  Removes a claim; recomputes contribution progress/status.
  Auth: Claim owner or event manager
```

## Notifications

```txt
GET /api/v1/notifications
  Query: unread? (bool), page?, per_page?
  Lists the caller's notifications, newest first.
  Auth: Recipient (self)

POST /api/v1/notifications/{notificationId}/read
  Marks one notification read (sets read_at).
  Auth: Recipient (self)

POST /api/v1/notifications/read-all
  Marks all of the caller's notifications read.
  Auth: Recipient (self)
```

## Device tokens

```txt
GET /api/v1/device-tokens
  Lists the caller's registered push tokens.
  Auth: Self

POST /api/v1/device-tokens
  Body: { token, platform }   platform ∈ ios | android | web
  Registers/updates a device push token.
  Auth: Self

DELETE /api/v1/device-tokens/{deviceTokenId}
  Removes a device push token.
  Auth: Owner (self)
```
