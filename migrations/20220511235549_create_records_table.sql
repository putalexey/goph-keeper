-- +goose Up
-- +goose StatementBegin
create table if not exists "records"
(
    "uuid"       uuid primary key,
    "user_uuid"  uuid         not null,
    "name"       varchar(512) not null,
    "type"       varchar(512) not null,
    "data"       bytea        null,
    "comment"    text         not null,
    "created_at" timestamp    null,
    "updated_at" timestamp    null,
    "deleted_at" timestamp    null
);
create unique index records_uniq_user_uuid_name_idx ON "records" ("user_uuid", "name");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "records";
-- +goose StatementEnd
