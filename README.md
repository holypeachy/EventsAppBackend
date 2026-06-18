# Events Backend

A small Go backend for planning private friend-group events. The app is built
around groups, invite-only event visibility, RSVPs, and lightweight participant
management so plans do not get buried in group chats.

## Goals

- Let users create and join private groups.
- Let group members create events and choose invited participants.
- Keep event visibility explicit: users only see events they participate in.
- RSVP status: `invited`, `going`, `maybe`, `declined`.
- Event roles: `owner`, `manager`, `participant`.

## Stack

- Go
- chi router
- PostgreSQL + pgx + goose (migrations)
- bcrypt password hashing
- JWT access tokens + Refresh tokens stored as hashes


## Configuration

Create `.env` file with content:

```env
DB_CONN=postgres://user:password@localhost:5432/events?sslmode=disable
JWT_SECRET=replace-me
```

The server currently listens on:

```txt
http://localhost:3000
```

## Database

Run migrations with goose:

```bash
goose -dir migrations postgres "$DB_CONN" up
```


## Running

```bash
go run .
```

Health check:

```txt
GET /api/v1/health
```

## API Docs

Frontend-facing endpoint documentation lives in:

```txt
docs/front_api.md
```

## Current Scope

Implemented MVP areas:

- Register, login, refresh, logout
- Create/list/join groups
- Group member listing and basic group management
- Create/list/update/delete events
- Participant-filtered event feeds
- RSVP
- Add/remove event participants
- Group and event authorization middleware

Future work:

- Event lifecycle/status rules
- Better logging with `log/slog`
- Tests for authorization and visibility rules
- Event contributions and claims
- Notifications
