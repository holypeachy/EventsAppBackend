# Frontend API Reference

Base URL: `/api/v1`

All request and response bodies are JSON unless noted otherwise.

Authenticated endpoints require:

```http
Authorization: Bearer <accessToken>
```

IDs are UUID strings. Timestamps are JSON strings produced from Go `time.Time`,
for example:

```json
"2026-06-18T18:30:00Z"
```

## Common Error Response

Most errors return:

```json
{
  "error": "message"
}
```

Common status codes:

```txt
400  malformed request, validation error, invalid path UUID
401  missing/invalid auth token or unauthorized operation
404  missing resource
409  duplicate/conflicting resource
422  database validation/constraint failure
500  internal server error
```

## Shared Schemas

### User

```json
{
  "id": "uuid",
  "username": "string",
  "email": "string",
  "createdAt": "timestamp"
}
```

### Login User

```json
{
  "id": "uuid",
  "username": "string",
  "email": "string"
}
```

### Group

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "createdBy": "uuid",
  "createdAt": "timestamp",
  "inviteCode": "string"
}
```

### Group Member

```json
{
  "groupId": "uuid",
  "userId": "uuid",
  "role": "member | admin | owner",
  "joinedAt": "timestamp"
}
```

### Event

```json
{
  "id": "uuid",
  "groupId": "uuid",
  "createdBy": "uuid",
  "name": "string",
  "description": "string",
  "location": "string",
  "status": "rsvp_open | rsvp_closed | cancelled | completed",
  "createdAt": "timestamp",
  "rsvpDeadline": "timestamp",
  "startsAt": "timestamp",
  "endsAt": "timestamp"
}
```

### Event Participant

```json
{
  "eventId": "uuid",
  "userId": "uuid",
  "status": "invited | going | maybe | declined",
  "role": "owner | manager | participant",
  "createdAt": "timestamp",
  "respondedAt": "timestamp",
  "username": "string"
}
```

### Status Response

```json
{
  "status": "string"
}
```

## Health

### GET `/health`

Auth: none

Response `200`:

```json
{
  "status": "ok"
}
```

## Auth

### POST `/auth/register`

Auth: none

Request:

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

Response `201`:

```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "user": {
    "id": "uuid",
    "username": "string",
    "email": "string"
  }
}
```

### POST `/auth/login`

Auth: none

Request:

```json
{
  "email": "string",
  "password": "string"
}
```

Response `200`:

```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "user": {
    "id": "uuid",
    "username": "string",
    "email": "string"
  }
}
```

### POST `/auth/refresh`

Auth: none

Request:

```json
{
  "refreshToken": "string"
}
```

Response `200`:

```json
{
  "accessToken": "string"
}
```

### POST `/auth/logout`

Auth: none

Request:

```json
{
  "refreshToken": "string"
}
```

Response `200`:

```json
{
  "status": "user successfully logged out"
}
```

## Groups

### POST `/groups`

Auth: required

Creates a group. The authenticated user becomes the group owner.

Request:

```json
{
  "name": "string",
  "description": "string"
}
```

Response `201`: `Group`

### GET `/groups`

Auth: required

Returns groups the authenticated user belongs to.

Response `200`:

```json
[
  {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "createdBy": "uuid",
    "createdAt": "timestamp",
    "inviteCode": "string"
  }
]
```

Empty response:

```json
[]
```

### POST `/groups/join`

Auth: required

Adds the authenticated user to a group by invite code.

Request:

```json
{
  "inviteCode": "string"
}
```

Response `200`: `Group`

### GET `/groups/{groupId}`

Auth: group member

Path params:

```txt
groupId: uuid
```

Response `200`: `Group`

### GET `/groups/{groupId}/members`

Auth: group member

Path params:

```txt
groupId: uuid
```

Response `200`:

```json
[
  {
    "id": "uuid",
    "username": "string",
    "email": "string",
    "createdAt": "timestamp"
  }
]
```

### POST `/groups/{groupId}/invite-code/regen`

Auth: group admin or owner

Path params:

```txt
groupId: uuid
```

Response `200`: `Group`

### PATCH `/groups/{groupId}`

Auth: group admin or owner

Path params:

```txt
groupId: uuid
```

Request:

```json
{
  "name": "string",
  "description": "string"
}
```

Response `200`: `Group`

### PATCH `/groups/{groupId}/members/{userId}`

Auth: group admin or owner

Current validation accepts only `member` or `admin`.

Path params:

```txt
groupId: uuid
userId: uuid
```

Request:

```json
{
  "role": "member | admin"
}
```

Response `200`: `Group Member`

### DELETE `/groups/{groupId}/members/{userId}`

Auth: group admin or owner

Path params:

```txt
groupId: uuid
userId: uuid
```

Response `200`:

```json
{
  "status": "member removed from group"
}
```

### DELETE `/groups/{groupId}`

Auth: group owner

Path params:

```txt
groupId: uuid
```

Response `200`:

```json
{
  "status": "group removed"
}
```

## Events

### GET `/events`

Auth: required

Returns all events the authenticated user participates in.

Response `200`:

```json
[
  {
    "id": "uuid",
    "groupId": "uuid",
    "createdBy": "uuid",
    "name": "string",
    "description": "string",
    "location": "string",
    "status": "rsvp_open | rsvp_closed | cancelled | completed",
    "createdAt": "timestamp",
    "rsvpDeadline": "timestamp",
    "startsAt": "timestamp",
    "endsAt": "timestamp"
  }
]
```

### POST `/groups/{groupId}/events`

Auth: group member

Creates an event in a group. The creator is added as event `owner`; valid
group-member `participantIds` are added as invited event participants. Invalid,
duplicate, or creator IDs are skipped.

Path params:

```txt
groupId: uuid
```

Request:

```json
{
  "name": "string",
  "description": "string",
  "location": "string",
  "status": "rsvp_open | rsvp_closed | cancelled | completed",
  "rsvpDeadline": "timestamp",
  "startsAt": "timestamp",
  "endsAt": "timestamp",
  "participantIds": ["uuid"]
}
```

Notes:

- `status` may be omitted or empty; the API defaults it to `rsvp_open`.
- `name`, `rsvpDeadline`, `startsAt`, and `endsAt` are required.

Response `201`: `Event`

### GET `/groups/{groupId}/events`

Auth: group member

Returns this group's events that the authenticated user participates in.

Path params:

```txt
groupId: uuid
```

Response `200`: array of `Event`

### GET `/events/{eventId}`

Auth: event participant

Path params:

```txt
eventId: uuid
```

Response `200`: `Event`

### PATCH `/events/{eventId}`

Auth: event manager or owner

Updates the full event object. Current API expects all fields, not a partial
patch.

Path params:

```txt
eventId: uuid
```

Request:

```json
{
  "name": "string",
  "description": "string",
  "location": "string",
  "status": "rsvp_open | rsvp_closed | cancelled",
  "rsvpDeadline": "timestamp",
  "startsAt": "timestamp",
  "endsAt": "timestamp"
}
```

Response `200`: `Event`

### DELETE `/events/{eventId}`

Auth: event manager or owner

Path params:

```txt
eventId: uuid
```

Response `200`:

```json
{
  "status": "event deleted"
}
```

## Event Participants

### GET `/events/{eventId}/participants`

Auth: event participant

Path params:

```txt
eventId: uuid
```

Response `200`:

```json
[
  {
    "eventId": "uuid",
    "userId": "uuid",
    "status": "invited | going | maybe | declined",
    "role": "owner | manager | participant",
    "createdAt": "timestamp",
    "respondedAt": "timestamp",
    "username": "string"
  }
]
```

### POST `/events/{eventId}/participants`

Auth: event manager or owner

Adds more group members to an event. Invalid, duplicate, and non-group-member
IDs are skipped.

Path params:

```txt
eventId: uuid
```

Request:

```json
{
  "participantIds": ["uuid"]
}
```

Response `200`:

```json
{
  "status": "event participants added"
}
```

### DELETE `/events/{eventId}/participants/{userId}`

Auth: event manager or owner

Removes an event participant. The API currently rejects removing event managers
and event owners.

Path params:

```txt
eventId: uuid
userId: uuid
```

Response `200`:

```json
{
  "status": "user removed from event"
}
```

### PATCH `/events/{eventId}/participants/{userId}/rsvp`

Auth: event participant, self only

The `{userId}` path parameter must match the authenticated user.

Path params:

```txt
eventId: uuid
userId: uuid
```

Request:

```json
{
  "status": "going | maybe | declined"
}
```

Response `200`: `Event Participant`

