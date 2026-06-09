# Friend Event Planner — Database Notes

A group-based event manager. A user creates a **group**; inside it they post
**events** and explicitly pick which members are invited. Uninvited members are
not notified and cannot see the event in their feed. Invited users RSVP
(going / maybe / declined). Each event has a **contributions** list (money
funds, food, items) that event managers set up; members claim parts of it
(e.g. a "$40 pizza fund" split across several people).

---

## Conventions

```txt
Engine        PostgreSQL
Primary keys  id UUID PRIMARY KEY DEFAULT gen_random_uuid()
Timestamps    TIMESTAMPTZ, stored in UTC
              created_at / updated_at default now()
Enums         native PostgreSQL ENUM types (see "Enum types")
Text identity citext for users.username and users.email
              (case-insensitive uniqueness)
```

- `gen_random_uuid()` is built in on PostgreSQL 13+ (no extension needed).
- Enable the `citext` extension: `CREATE EXTENSION IF NOT EXISTS citext;`.
- Every `*_at` column is `TIMESTAMPTZ`; clients send/receive UTC.

---

## Enum types

```txt
group_member_role   member | admin | owner
event_status        scheduled | cancelled | completed
participant_status  invited | going | maybe | declined
participant_role    owner | manager | participant
contribution_type   money | food | item | other
contribution_status open | claimed | fulfilled | cancelled
notification_type   group_added | event_invited | event_updated
                    | event_cancelled | contribution_added
                    | contribution_claimed | contribution_fulfilled
                    | rsvp_reminder
device_platform     ios | android | web
```

---

## users

Fields:

```txt
id            uuid          PK, default gen_random_uuid()
username      citext        NOT NULL, UNIQUE
email         citext        NOT NULL, UNIQUE
password_hash text          NOT NULL
created_at    timestamptz   NOT NULL, default now()
updated_at    timestamptz   NOT NULL, default now()
```

Notes:

- Store bcrypt password hashes only.
- Do not store tokens directly on the user table.
- `citext` makes username/email uniqueness case-insensitive.

---

## groups

Fields:

```txt
id          uuid         PK, default gen_random_uuid()
name        text         NOT NULL
created_by  uuid         NOT NULL
created_at  timestamptz  NOT NULL, default now()
updated_at  timestamptz  NOT NULL, default now()
```

Relationships:

```txt
created_by → users.id   ON DELETE RESTRICT
```

Notes:

- A group is created by a user.
- On creation the creator also gets a `group_members` row with
  `role = 'owner'` (application-enforced).

---

## group_members

Fields:

```txt
group_id   uuid               NOT NULL
user_id    uuid               NOT NULL
role       group_member_role  NOT NULL, default 'member'
joined_at  timestamptz        NOT NULL, default now()

PRIMARY KEY (group_id, user_id)
```

Relationships:

```txt
group_id → groups.id   ON DELETE CASCADE
user_id  → users.id    ON DELETE CASCADE
```

Notes:

- Replaces separate admin tables. Authorization can use `role` directly.
- Composite PK `(group_id, user_id)` enforces one membership row per user
  per group.

---

## events

Fields:

```txt
id           uuid          PK, default gen_random_uuid()
group_id     uuid          NOT NULL
created_by   uuid          NOT NULL
name         text          NOT NULL
description  text          NULL
location     text          NULL
starts_at    timestamptz   NOT NULL
ends_at      timestamptz   NULL
status       event_status  NOT NULL, default 'scheduled'
created_at   timestamptz   NOT NULL, default now()
updated_at   timestamptz   NOT NULL, default now()

CHECK (ends_at IS NULL OR ends_at >= starts_at)
```

Relationships:

```txt
group_id   → groups.id   ON DELETE CASCADE
created_by → users.id    ON DELETE RESTRICT
```

Notes:

- Every event belongs to a group.
- **Visibility is per-event, not per-group.** A user can see an event only
  if a row exists for them in `event_participants`. There is **no exclusions
  table** — the client sends the chosen invite list at creation, and excluded
  members simply get no `event_participants` row.
- A user's feed = the events they have an `event_participants` row for.

---

## event_participants

Fields:

```txt
event_id      uuid                NOT NULL
user_id       uuid                NOT NULL
status        participant_status  NOT NULL, default 'invited'
role          participant_role    NOT NULL, default 'participant'
responded_at  timestamptz         NULL
created_at    timestamptz         NOT NULL, default now()

PRIMARY KEY (event_id, user_id)
```

Relationships:

```txt
event_id → events.id   ON DELETE CASCADE
user_id  → users.id    ON DELETE CASCADE
```

Notes:

- This single table is the **invite list**, the **visibility list**, the
  **RSVP store**, and the **per-event role list** (`role = 'owner'`,
  `'manager'`, or `'participant'`).
- Rows exist only for invited users.
- `responded_at` is set when the user changes `status` away from `invited`.

---

## contributions

Fields:

```txt
id                  uuid                 PK, default gen_random_uuid()
event_id            uuid                 NOT NULL
created_by          uuid                 NOT NULL
name                text                 NOT NULL
description         text                 NULL
type                contribution_type    NOT NULL
status              contribution_status  NOT NULL, default 'open'
target_amount_cents bigint               NULL   -- when type = 'money'
target_quantity     int                  NULL   -- when type in
                                                 -- (food,item,other)
created_at          timestamptz          NOT NULL, default now()
updated_at          timestamptz          NOT NULL, default now()
```

Relationships:

```txt
event_id   → events.id   ON DELETE CASCADE
created_by → users.id    ON DELETE RESTRICT
```

Notes:

- `target_amount_cents` (renamed from `amount_cents`) and `target_quantity`
  (renamed from `quantity_needed`) describe the goal.
- `claimed_by` and `quantity_claimed` were **removed**. Progress is derived
  by summing `contribution_claims` (see below): `SUM(amount_cents)` for money,
  `SUM(quantity)` for food/item/other.
- Only event managers (event owner or a participant with `role = 'manager'`)
  may create/edit/cancel contributions (application-enforced).

Examples:

- $40 pizza fund (`type = money`, `target_amount_cents = 4000`)
- drinks (`type = food`)
- chips (`type = food`)
- plates/cups (`type = item`)

---

## contribution_claims

Each row is one person pledging part (or all) of a contribution. A single
contribution can have many claims (e.g. a money pool split three ways).

Fields:

```txt
id              uuid         PK, default gen_random_uuid()
contribution_id uuid         NOT NULL
user_id         uuid         NOT NULL
amount_cents    bigint       NULL   -- when contribution.type = 'money'
quantity        int          NULL   -- when food/item/other
note            text         NULL   -- e.g. "bringing pepperoni"
claimed_at      timestamptz  NOT NULL, default now()

CHECK (amount_cents IS NOT NULL OR quantity IS NOT NULL)
CHECK (amount_cents IS NULL OR amount_cents > 0)
CHECK (quantity     IS NULL OR quantity     > 0)
```

Relationships:

```txt
contribution_id → contributions.id   ON DELETE CASCADE
user_id         → users.id           ON DELETE CASCADE
```

Notes:

- **No unique `(contribution_id, user_id)`** — a user may pledge more than
  once into the same pool. Progress is the `SUM` over all claims.
- Only invited participants of the contribution's event may insert a claim
  (cross-table rule, application-enforced).
- Contribution `status` transitions `open → claimed → fulfilled` are managed
  by the application as the summed claims reach the target.

---

## notifications

Fields:

```txt
id              uuid               PK, default gen_random_uuid()
user_id         uuid               NOT NULL   -- recipient
type            notification_type  NOT NULL
event_id        uuid               NULL
group_id        uuid               NULL
contribution_id uuid               NULL
message         text               NOT NULL
read_at         timestamptz        NULL
created_at      timestamptz        NOT NULL, default now()
```

Relationships:

```txt
user_id         → users.id          ON DELETE CASCADE
event_id        → events.id         ON DELETE CASCADE
group_id        → groups.id         ON DELETE CASCADE
contribution_id → contributions.id  ON DELETE CASCADE
```

Notes:

- Stores in-app notifications.
- `event_id` is now nullable; `group_id` and `contribution_id` were added so
  group-level and contribution-level notifications can be represented.
- `type` uses the `notification_type` enum.

---

## refresh_tokens

Fields:

```txt
id            uuid         PK, default gen_random_uuid()
user_id       uuid         NOT NULL
token_hash    text         NOT NULL, UNIQUE
expires_at    timestamptz  NOT NULL
revoked_at    timestamptz  NULL
last_used_at  timestamptz  NULL
created_at    timestamptz  NOT NULL, default now()
```

Relationships:

```txt
user_id → users.id   ON DELETE CASCADE
```

Notes:

- Used for sliding login sessions.
- Store token hashes, not raw tokens. `token_hash` is unique.

---

## Optional Later

## device_tokens

Fields:

```txt
id          uuid             PK, default gen_random_uuid()
user_id     uuid             NOT NULL
token       text             NOT NULL, UNIQUE
platform    device_platform  NOT NULL
created_at  timestamptz      NOT NULL, default now()
updated_at  timestamptz      NOT NULL, default now()
```

Relationships:

```txt
user_id → users.id   ON DELETE CASCADE
```

Notes:

- Used for push notifications later.
- `platform` enum: `ios | android | web`.

---

## Constraints & indexes

Uniqueness / keys:

```txt
users.username                       UNIQUE
users.email                          UNIQUE
group_members                        PK (group_id, user_id)
event_participants                   PK (event_id, user_id)
refresh_tokens.token_hash            UNIQUE
device_tokens.token                  UNIQUE
```

Indexes (query performance):

```txt
events(group_id)
events(starts_at)
event_participants(user_id)              -- per-user feed query
contribution_claims(contribution_id)     -- sum progress for a contribution
contribution_claims(user_id)             -- a user's claims
notifications(user_id, read_at)          -- unread list
refresh_tokens(user_id)
```

Foreign-key delete behavior:

```txt
RESTRICT  groups.created_by, events.created_by, contributions.created_by
          (keep authorship; block deleting a referenced user)
CASCADE   all membership / participation / child rows
          (group_members, event_participants, contributions,
           contribution_claims, notifications, refresh_tokens,
           device_tokens)
```

---

## Behavioral rules / flows

- **Feed / visibility.** A user sees an event iff an `event_participants` row
  exists for `(event, user)`. No exclusions table; the client controls who is
  added at creation.

- **Create event.** Insert the event → insert an `event_participants` row for
  the creator with `role = 'owner'` → insert `event_participants` rows
  (`status = 'invited'`, `role = 'participant'`) for each other selected user
  → enqueue `event_invited` notifications for the invited users.

- **Event managers.** The event owner is a manager. A manager may promote
  any participant of that event to manager by setting
  `event_participants.role = 'manager'` (and demote back to `'participant'`).
  Managers (`role = 'owner'` or `role = 'manager'`) have equal powers:
  edit/cancel the event, manage the invite list (add/remove participants),
  and create/edit/cancel contributions. All application-enforced.

- **RSVP.** An invited user sets `event_participants.status` to
  `going` / `maybe` / `declined` and `responded_at` is set.

- **Cancel event.** A manager sets `events.status = 'cancelled'`; notify
  participants with `event_cancelled`.

- **Manage contributions.** Only event managers may create/edit/cancel
  contributions.

- **Claim.** Only invited participants of the contribution's event may insert
  a `contribution_claims` row. Progress = `SUM(amount_cents)` (money) or
  `SUM(quantity)` (food/item/other) vs the contribution's target; the
  application transitions `contribution.status` open → claimed → fulfilled.
