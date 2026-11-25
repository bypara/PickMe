-- +goose Up
CREATE TABLE animals (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL,
    found_location text NOT NULL,
    breed text NOT NULL,
    created_by bigint NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT animals_name_length CHECK (
        char_length(name) BETWEEN 1 AND 100 AND name !~ '^[[:space:]]|[[:space:]]$'
    ),
    CONSTRAINT animals_found_location_length CHECK (
        char_length(found_location) BETWEEN 1 AND 255 AND found_location !~ '^[[:space:]]|[[:space:]]$'
    ),
    CONSTRAINT animals_breed_length CHECK (
        char_length(breed) BETWEEN 1 AND 100 AND breed !~ '^[[:space:]]|[[:space:]]$'
    )
);

CREATE INDEX animals_catalog_order_idx ON animals (created_at DESC, id DESC);

CREATE TABLE animal_photos (
    animal_id bigint PRIMARY KEY REFERENCES animals (id) ON DELETE CASCADE,
    content_type text NOT NULL,
    data bytea NOT NULL,
    CONSTRAINT animal_photos_content_type CHECK (content_type IN ('image/jpeg', 'image/png')),
    CONSTRAINT animal_photos_data_size CHECK (octet_length(data) BETWEEN 1 AND 5 * 1024 * 1024)
);

-- +goose Down
DROP TABLE animal_photos;
DROP TABLE animals;
