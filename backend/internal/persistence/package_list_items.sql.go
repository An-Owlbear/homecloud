// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: package_list_items.sql

package persistence

import (
	"context"
	"database/sql"
)

const getCategories = `-- name: GetCategories :many
SELECT category FROM package_categories
`

func (q *Queries) GetCategories(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		items = append(items, category)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const addCategory = `-- name: addCategory :exec

INSERT OR IGNORE INTO package_categories (category)
VALUES (?1)
`

// -- name: getPackageListItems :many
// SELECT sqlc.embed(package_list_items), string_agg(package_categories.name, ',')
// FROM package_list_items
// INNER JOIN package_categories ON package_list_items.id = package_categories.package_id
// GROUP BY package_list_items.id;
//
// functions used for the process of adding packages to the DB
func (q *Queries) addCategory(ctx context.Context, category string) error {
	_, err := q.db.ExecContext(ctx, addCategory, category)
	return err
}

const addPackageCategoryDefinition = `-- name: addPackageCategoryDefinition :exec
INSERT OR IGNORE INTO package_category_definitions (category, package_id)
VALUES (?1, ?2)
`

type addPackageCategoryDefinitionParams struct {
	Category  string `json:"category"`
	PackageID string `json:"package_id"`
}

func (q *Queries) addPackageCategoryDefinition(ctx context.Context, arg addPackageCategoryDefinitionParams) error {
	_, err := q.db.ExecContext(ctx, addPackageCategoryDefinition, arg.Category, arg.PackageID)
	return err
}

const deletePackageCategoryDefinitions = `-- name: deletePackageCategoryDefinitions :exec
DELETE FROM package_category_definitions
WHERE package_id = ?1
`

func (q *Queries) deletePackageCategoryDefinitions(ctx context.Context, packageID string) error {
	_, err := q.db.ExecContext(ctx, deletePackageCategoryDefinitions, packageID)
	return err
}

const getPackageCategoryDefinitions = `-- name: getPackageCategoryDefinitions :many
SELECT category FROM package_category_definitions
WHERE package_id = ?1
`

func (q *Queries) getPackageCategoryDefinitions(ctx context.Context, packageID string) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getPackageCategoryDefinitions, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		items = append(items, category)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPackageListItem = `-- name: getPackageListItem :one
SELECT package_list_items.id, package_list_items.name, package_list_items.version, package_list_items.author, package_list_items.description, package_list_items.image_url, category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
WHERE package_list_items.id = ?1
`

type getPackageListItemRow struct {
	PackageListItem PackageListItem `json:"package_list_item"`
	Category        sql.NullString  `json:"category"`
	Installed       bool            `json:"installed"`
}

func (q *Queries) getPackageListItem(ctx context.Context, id string) (getPackageListItemRow, error) {
	row := q.db.QueryRowContext(ctx, getPackageListItem, id)
	var i getPackageListItemRow
	err := row.Scan(
		&i.PackageListItem.ID,
		&i.PackageListItem.Name,
		&i.PackageListItem.Version,
		&i.PackageListItem.Author,
		&i.PackageListItem.Description,
		&i.PackageListItem.ImageUrl,
		&i.Category,
		&i.Installed,
	)
	return i, err
}

const getPackageListItems = `-- name: getPackageListItems :many

SELECT package_list_items.id, package_list_items.name, package_list_items.version, package_list_items.author, package_list_items.description, package_list_items.image_url, category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
ORDER BY package_list_items.id
`

type getPackageListItemsRow struct {
	PackageListItem PackageListItem `json:"package_list_item"`
	Category        sql.NullString  `json:"category"`
	Installed       bool            `json:"installed"`
}

// functions relating to retrieving packages
func (q *Queries) getPackageListItems(ctx context.Context) ([]getPackageListItemsRow, error) {
	rows, err := q.db.QueryContext(ctx, getPackageListItems)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []getPackageListItemsRow
	for rows.Next() {
		var i getPackageListItemsRow
		if err := rows.Scan(
			&i.PackageListItem.ID,
			&i.PackageListItem.Name,
			&i.PackageListItem.Version,
			&i.PackageListItem.Author,
			&i.PackageListItem.Description,
			&i.PackageListItem.ImageUrl,
			&i.Category,
			&i.Installed,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const searchPackageListItems = `-- name: searchPackageListItems :many
WITH has_category AS (
    SELECT package_id, MAX(CASE WHEN (?3 = '' OR category = ?3) THEN 1 ELSE 0 END) AS has_category
    FROM package_category_definitions
    GROUP BY package_id
)
SELECT package_list_items.id, package_list_items.name, package_list_items.version, package_list_items.author, package_list_items.description, package_list_items.image_url, category, CAST(CASE WHEN apps.id IS NOT NULL THEN TRUE ELSE FALSE END AS BOOLEAN) AS installed
FROM package_list_items
LEFT JOIN has_category ON package_list_items.id = has_category.package_id
LEFT JOIN package_category_definitions ON package_list_items.id = package_category_definitions.package_id
LEFT JOIN apps ON package_list_items.id = apps.id
WHERE (?1 = '' OR lower(package_list_items.id) LIKE '%' || lower(?1) || '%')
AND (?2 = '' OR author = ?2)
AND has_category = 1
ORDER BY package_list_items.id
`

type searchPackageListItemsParams struct {
	ID       interface{} `json:"id"`
	Author   interface{} `json:"author"`
	Category interface{} `json:"category"`
}

type searchPackageListItemsRow struct {
	PackageListItem PackageListItem `json:"package_list_item"`
	Category        sql.NullString  `json:"category"`
	Installed       bool            `json:"installed"`
}

func (q *Queries) searchPackageListItems(ctx context.Context, arg searchPackageListItemsParams) ([]searchPackageListItemsRow, error) {
	rows, err := q.db.QueryContext(ctx, searchPackageListItems, arg.ID, arg.Author, arg.Category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []searchPackageListItemsRow
	for rows.Next() {
		var i searchPackageListItemsRow
		if err := rows.Scan(
			&i.PackageListItem.ID,
			&i.PackageListItem.Name,
			&i.PackageListItem.Version,
			&i.PackageListItem.Author,
			&i.PackageListItem.Description,
			&i.PackageListItem.ImageUrl,
			&i.Category,
			&i.Installed,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const writePackageListItem = `-- name: writePackageListItem :exec
INSERT OR REPLACE INTO package_list_items (id, name, version, author, description, image_url)
VALUES (?1, ?2, ?3, ?4, ?5, ?6)
`

type writePackageListItemParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
}

func (q *Queries) writePackageListItem(ctx context.Context, arg writePackageListItemParams) error {
	_, err := q.db.ExecContext(ctx, writePackageListItem,
		arg.ID,
		arg.Name,
		arg.Version,
		arg.Author,
		arg.Description,
		arg.ImageUrl,
	)
	return err
}
