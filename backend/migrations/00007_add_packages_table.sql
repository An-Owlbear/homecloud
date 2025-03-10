-- +goose Up
-- +goose StatementBegin
CREATE TABLE package_list_items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT NOT NULL,
    image_url TEXT NOT NULL
);

CREATE TABLE package_categories (
    category TEXT PRIMARY KEY
);

CREATE TABLE package_category_definitions (
    package_id TEXT NOT NULL,
    category TEXT NOT NULL,
    PRIMARY KEY (category, package_id),
    FOREIGN KEY (category) REFERENCES package_categories (category),
    FOREIGN KEY (package_id) REFERENCES package_list_items (id)
);
CREATE INDEX idx_package_category_definitions_name ON package_category_definitions(category);
CREATE INDEX idx_package_category_definitions_package_id ON package_category_definitions(package_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE package_category_definitions;
DROP TABLE package_categories;
DROP TABLE package_list_items;
-- +goose StatementEnd
