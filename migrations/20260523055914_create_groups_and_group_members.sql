-- +goose Up
CREATE TABLE groups(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    description TEXT,

    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    invite_code TEXT UNIQUE NOT NULL,

    CONSTRAINT fk_groups_user
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_groups_created_by
ON groups(created_by);


CREATE TYPE group_role AS ENUM (
    'member',
    'admin',
    'owner'
);

CREATE TABLE group_members(
    group_id UUID NOT NULL,
    user_id UUID NOT NULL,

    role group_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (group_id, user_id),

    CONSTRAINT fk_group_members_group
        FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_group_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
CREATE INDEX idx_group_members_user_id
ON group_members(user_id);


-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TYPE IF EXISTS group_role;
