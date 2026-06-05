-- +goose Up

CREATE TYPE user_role AS ENUM ('user', 'operator', 'organizer', 'administrator');

CREATE TABLE users (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login                 TEXT UNIQUE,
    password              TEXT,
    last_name             TEXT NOT NULL DEFAULT '',
    first_name            TEXT NOT NULL DEFAULT '',
    middle_name           TEXT NOT NULL DEFAULT '',
    registered            BOOLEAN NOT NULL DEFAULT FALSE,
    active                BOOLEAN NOT NULL DEFAULT FALSE,
    role                  user_role NOT NULL DEFAULT 'user',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    last_online           TIMESTAMPTZ,
    auth_provider         TEXT NOT NULL DEFAULT 'local',
    language              TEXT NOT NULL DEFAULT 'ru',
    notification_language TEXT NOT NULL DEFAULT 'ru'
);

-- +goose Down

DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
