CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number     TEXT UNIQUE NOT NULL,
    display_name     TEXT,
    about            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_devices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name   TEXT,
    device_type   TEXT,
    public_key    TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_group        BOOLEAN NOT NULL DEFAULT FALSE,
    title           TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT DEFAULT 'member', -- admin, owner, member
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE messages (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id  UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    content          TEXT,
    message_type     TEXT NOT NULL DEFAULT 'text',
    reply_to_id      UUID REFERENCES messages(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE media (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id     UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    file_url       TEXT NOT NULL,
    mime_type      TEXT NOT NULL,
    file_size      BIGINT,
    uploaded_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE message_receipts (
    message_id      UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    PRIMARY KEY (message_id, user_id)
);

CREATE TABLE message_reactions (
    message_id   UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction     TEXT NOT NULL,
    reacted_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, user_id)
);

CREATE TABLE contacts (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_name  TEXT,
    PRIMARY KEY (user_id, contact_id)
);

CREATE TABLE encryption_keys (
    device_id        UUID PRIMARY KEY REFERENCES user_devices(id) ON DELETE CASCADE,
    identity_key     TEXT NOT NULL,
    prekey_public    TEXT NOT NULL,
    prekey_signature TEXT NOT NULL
);