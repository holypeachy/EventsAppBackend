-- +goose Up
CREATE TYPE event_status AS ENUM (
    'rsvp_open',
    'rsvp_closed',
    'cancelled',
    'completed'
);
CREATE TYPE participant_status AS ENUM (
    'invited',
    'going',
    'maybe',
    'declined'
);
CREATE TYPE participant_role AS ENUM (
    'owner',
    'manager',
    'participant'
);

CREATE TABLE events(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    group_id UUID NOT NULL,
    created_by UUID,

    name TEXT NOT NULL,
    description TEXT,
    location TEXT,
    status event_status NOT NULL DEFAULT 'rsvp_open',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rsvp_deadline TIMESTAMPTZ NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    

    CONSTRAINT fk_events_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_events_user
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_starts_deadline
        CHECK (rsvp_deadline <= starts_at),
    CONSTRAINT chk_ends_starts
        CHECK (ends_at >= starts_at)
);

CREATE TABLE event_participants(
    event_id UUID NOT NULL,
    user_id UUID NOT NULL,

    status participant_status NOT NULL DEFAULT 'invited',
    role participant_role NOT NULL DEFAULT 'participant',
   
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMPTZ,

    PRIMARY KEY (event_id, user_id),

    CONSTRAINT fk_event_participants_event
        FOREIGN KEY (event_id)
        REFERENCES events(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_event_participants_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS event_participants;
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS participant_role;
DROP TYPE IF EXISTS participant_status;
DROP TYPE IF EXISTS event_status;
