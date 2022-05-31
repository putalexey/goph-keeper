-- +goose Up
-- +goose StatementBegin
create table if not exists "users"
(
    "uuid"     uuid primary key,
    "login"    varchar(512) not null,
    "password" varchar(512) not null
);
create unique index users_uniq_login_idx ON "users" ("login");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "users";
-- +goose StatementEnd
