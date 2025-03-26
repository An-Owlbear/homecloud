-- -- name: getPackageListItems :many
-- SELECT sqlc.embed(package_list_items), string_agg(package_categories.name, ',')
-- FROM package_list_items
-- INNER JOIN package_categories ON package_list_items.id = package_categories.package_id
-- GROUP BY package_list_items.id;

--
-- functions used for the process of adding packages to the DB
--
-- name: addCategory :exec
INSERT OR IGNORE INTO package_categories (category)
VALUES (sqlc.arg(category));

-- name: writePackageListItem :exec
INSERT OR REPLACE INTO package_list_items (id, name, version, author, description, image_url)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(version), sqlc.arg(author), sqlc.arg(description), sqlc.arg(image_url));

-- name: addPackageCategoryDefinition :exec
INSERT OR IGNORE INTO package_category_definitions (category, package_id)
VALUES (sqlc.arg(category), sqlc.arg(package_id));

-- name: getPackageCategoryDefinitions :many
SELECT category FROM package_category_definitions
WHERE package_id = sqlc.arg(package_id);

-- name: deletePackageCategoryDefinitions :exec
DELETE FROM package_category_definitions
WHERE package_id = sqlc.arg(package_id);

--
-- functions relating to retrieving packages
--

-- name: getPackageListItems :many
SELECT sqlc.embed(package_list_items), category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
ORDER BY package_list_items.id;

-- name: getPackageListItem :one
SELECT sqlc.embed(package_list_items), category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
WHERE package_list_items.id = sqlc.arg(id);

-- name: searchPackageListItems :many
WITH has_category AS (
    SELECT package_id, MAX(CASE WHEN (sqlc.arg(category) = '' OR category = sqlc.arg(category)) THEN 1 ELSE 0 END) AS has_category
    FROM package_category_definitions
    GROUP BY package_id
)
SELECT sqlc.embed(package_list_items), category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN has_category ON package_list_items.id = has_category.package_id
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
WHERE (sqlc.arg(id) = '' OR lower(package_list_items.id) LIKE '%' || lower(sqlc.arg(id)) || '%')
AND (sqlc.arg(author) = '' OR author = sqlc.arg(author))
AND has_category = 1
ORDER BY package_list_items.id;

-- name: GetCategories :many
SELECT category FROM package_categories;

-- name: GetPopularCategories :many
SELECT category FROM package_category_definitions
GROUP BY category
ORDER BY count(category)
LIMIT 6;

-- name: GetNewPackages :many
SELECT * FROM package_list_items
LIMIT 10;