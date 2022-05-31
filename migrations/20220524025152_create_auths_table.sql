-- +goose Up
-- +goose StatementBegin
create table if not exists "auths"
(
    "uuid"       uuid primary key,
    "user_uuid"  uuid         not null,
    "token"      varchar(512) not null,
    "created_at" timestamp    null
);
create unique index auths_user_uuid_idx ON "auths" ("user_uuid", "token");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "auths";
-- +goose StatementEnd
