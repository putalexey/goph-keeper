-- +goose Up
-- +goose StatementBegin
ALTER TABLE "records"
    ADD COLUMN "data_encrypted" bool not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "records"
    DROP COLUMN "data_encrypted";
-- +goose StatementEnd
