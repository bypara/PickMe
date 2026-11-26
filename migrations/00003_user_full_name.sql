-- +goose Up
-- Ранее созданные аккаунты могут войти и без имени. Для новых имя проверяет регистрация.
ALTER TABLE users ADD COLUMN full_name text;

ALTER TABLE users ADD CONSTRAINT users_full_name_length CHECK (
    char_length(full_name) BETWEEN 2 AND 100 AND full_name !~ '^[[:space:]]|[[:space:]]$'
);

-- +goose Down
ALTER TABLE users DROP COLUMN full_name;
