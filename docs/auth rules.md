# Friend Event Planner — Authorization Rules

Companion to `friend_event_planner_db_notes_trimmed.md` and
`friend_event_planner_rest_api.md`. Defines **who may do what**. Every
endpoint in the REST API doc maps to exactly one row in the
[Per-endpoint authorization](#4-per-endpoint-authorization) table.

Auth scheme: **JWT bearer access token + rotating refresh token**
(`Authorization: Bearer <access>`). Refresh/rotation backed by the
`refresh_tokens` table.

---

## 1. Principals / roles

| Principal | Definition (source) |
|---|---|
| **Anonymous** | No / invalid access token. |
| **Authenticated user** | Any valid access token. |
| **Group member** | `group_members` row with `role='member'` for the group. |
| **Group admin** | `group_members.role='admin'`. |
| **Group owner** | `group_members.role='owner'` (the creator; one per group). |
| **Event participant** | Any `event_participants` row for the event (any `role`/`status`). |
| **Event manager** | Event `created_by` **OR** `event_participants.role='admin'` for that event. |
| **Claim owner** | `contribution_claims.user_id` = the requester. |
| **Self** | The authenticated user acting on their own user / notifications / device-tokens. |

Notes:

- Group roles escalate: `owner` ⊇ `admin` ⊇ `member`.
- "Event manager" is independent of group role. The event creator is always a
  manager; a manager may promote any participant to manager
  (`event_participants.role='admin'`) and demote back.
- A user must be a member of a group to be invited to that group's events.

---

## 2. Capability matrix — group scope

| Capability | member | admin | owner |
|---|:--:|:--:|:--:|
| View group | Y | Y | Y |
| View member list | Y | Y | Y |
| Create event in group | Y | Y | Y |
| Edit group (name) | – | Y | Y |
| Add group member | – | Y | Y |
| Remove group member | – | Y¹ | Y |
| Change a member's role | – | – | Y |
| Delete group | – | – | Y |
| Leave group (self) | Y² | Y² | –³ |

¹ Admin may remove `member`s, not other admins/owner.
² Any non-owner may remove their own membership.
³ Owner must transfer ownership before leaving (owner role change is
owner-only).

## 2b. Capability matrix — event scope

| Capability | participant | event manager | claim owner |
|---|:--:|:--:|:--:|
| View event + sub-resources | Y | Y | Y⁴ |
| Edit event | – | Y | – |
| Cancel event | – | Y | – |
| View participant list | Y | Y | Y⁴ |
| Add participant (invite) | – | Y | – |
| Remove participant | self⁵ | Y | – |
| RSVP (going/maybe/declined) | self only | self only | self only |
| Promote/demote event manager | – | Y | – |
| Create/edit/cancel contribution | – | Y | – |
| View contributions | Y | Y | Y⁴ |
| Claim a contribution | Y⁶ | Y⁶ | Y⁶ |
| Edit/delete **own** claim | Y | Y | Y |
| Edit/delete **any** claim | – | Y | – |

⁴ A claim owner is by definition an event participant; listed for clarity.
⁵ A participant may remove only their own participation (leave the event); the
event creator cannot be removed.
⁶ Only **invited participants of the event** may create claims (cross-table
rule — see §5).

## 2c. Capability matrix — account scope

| Capability | self | anyone else |
|---|:--:|:--:|
| Read/update own profile (`users/me`) | Y | – |
| List authenticated users for lookup | Y⁷ | Y⁷ |
| Read/mark own notifications | Y | – |
| Register/list/delete own device tokens | Y | – |

⁷ Any authenticated user may search the user directory (limited fields:
id, username) to pick invitees. No one may read another user's email/profile.

---

## 3. Cross-cutting rules

- **Event visibility.** An event and all of its sub-resources
  (participants, contributions, claims) are accessible **only to its
  participants**. A user with no `event_participants` row for the event
  receives `404 Not Found` (existence is not disclosed). There is **no
  exclusions table** — exclusion = simply having no participant row.
- **Self-only writes.** A user may always read/update their own profile,
  notifications, and device tokens, and may never act on another user's.
- **RSVP is self-only.** `PATCH .../participants/{userId}/rsvp` requires
  `{userId}` == the authenticated user. Managers do **not** RSVP for others.
- **Role changes.** Only an event manager changes another participant's
  `event_participants.role`. Only a group owner changes
  `group_members.role`.
- **`RESTRICT` foreign keys.** A user referenced by `groups.created_by`,
  `events.created_by`, or `contributions.created_by` cannot be hard-deleted;
  such requests fail with `409 Conflict`.
- **State guards.** Mutations on a `cancelled` event (new contributions,
  claims, RSVP) are rejected with `409 Conflict`.
- **Status codes.** `401` missing/invalid token · `403` authenticated but
  not permitted · `404` hidden or missing resource · `409` conflict/state ·
  `422` validation.

---

## 4. Per-endpoint authorization

Mirrors the inventory in `friend_event_planner_rest_api.md` (one row each).

| Method & endpoint | Who may call |
|---|---|
| `POST /auth/register` | Anonymous |
| `POST /auth/login` | Anonymous |
| `POST /auth/refresh` | Bearer of a valid, non-revoked refresh token |
| `POST /auth/logout` | Owner of the supplied refresh token |
| `GET /users/me` | Self |
| `PATCH /users/me` | Self |
| `GET /users?query=` | Any authenticated user (limited fields) |
| `POST /groups` | Any authenticated user (becomes owner) |
| `GET /groups` | Any authenticated user (returns own memberships) |
| `GET /groups/{groupId}` | Group member |
| `PATCH /groups/{groupId}` | Group owner or admin |
| `DELETE /groups/{groupId}` | Group owner |
| `GET /groups/{groupId}/members` | Group member |
| `POST /groups/{groupId}/members` | Group owner or admin |
| `PATCH /groups/{groupId}/members/{userId}` | Group owner |
| `DELETE /groups/{groupId}/members/{userId}` | Group owner/admin, or self-leave |
| `POST /groups/{groupId}/events` | Group member |
| `GET /groups/{groupId}/events` | Group member (only events they participate in) |
| `GET /events` | Authenticated (only events they participate in) |
| `GET /events/{eventId}` | Event participant |
| `PATCH /events/{eventId}` | Event manager |
| `POST /events/{eventId}/cancel` | Event manager |
| `GET /events/{eventId}/participants` | Event participant |
| `POST /events/{eventId}/participants` | Event manager |
| `DELETE /events/{eventId}/participants/{userId}` | Event manager, or self-leave (not the creator) |
| `PATCH /events/{eventId}/participants/{userId}/rsvp` | Self only (must be that participant) |
| `PATCH /events/{eventId}/participants/{userId}/role` | Event manager |
| `POST /events/{eventId}/contributions` | Event manager |
| `GET /events/{eventId}/contributions` | Event participant |
| `GET /contributions/{contributionId}` | Participant of the contribution's event |
| `PATCH /contributions/{contributionId}` | Event manager |
| `DELETE /contributions/{contributionId}` | Event manager |
| `POST /contributions/{contributionId}/claims` | Invited participant of the contribution's event |
| `GET /contributions/{contributionId}/claims` | Participant of the contribution's event |
| `PATCH /contributions/{contributionId}/claims/{claimId}` | Claim owner or event manager |
| `DELETE /contributions/{contributionId}/claims/{claimId}` | Claim owner or event manager |
| `GET /notifications` | Recipient (self) |
| `POST /notifications/{notificationId}/read` | Recipient (self) |
| `POST /notifications/read-all` | Recipient (self) |
| `GET /device-tokens` | Self |
| `POST /device-tokens` | Self |
| `DELETE /device-tokens/{deviceTokenId}` | Owner (self) |
