-- +goose Up
-- +goose StatementBegin
create table if not exists "events"
(
    "uuid"        uuid primary key,
    "user_uuid"   uuid         not null,
    "record_uuid" uuid         not null,
    "date"        timestamp    not null,
    "action"      varchar(512) not null,
    "data"        bytea        null
);
create index events_user_date_idx ON "events" ("user_uuid", "date");
create index events_record_date_idx ON "events" ("record_uuid", "date");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists "events";
-- +goose StatementEnd
