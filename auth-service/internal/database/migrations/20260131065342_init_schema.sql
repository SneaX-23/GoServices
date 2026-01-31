-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,

    created_at TIMESTAMPZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hashed_token TEXT NOT NULL UNIQUE,

    user_id UUID NOT NULL,
    replaced_by UUID NULL,

    expires_at TIMESTAMPZ NOT NULL,
    created_at TIMESTAMPZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_refresh_tokens_user 
        FOREIGN KEY (user_id)
        REFERENCES users(id) 
        on DELETE cascade
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_hashed_token ON refresh_tokens(hashed_token);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
