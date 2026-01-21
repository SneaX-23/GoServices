-- +goose Up
-- +goose StatementBegin
CREATE table IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) not null,
    username varchar(255) not null
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users
-- +goose StatementEnd
