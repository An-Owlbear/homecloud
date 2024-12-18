package persistence

import (
	"context"
	"encoding/json"
)

const getApps = `-- name: GetApps :many
SELECT id, json(schema) as schema, date_added FROM apps
`

type GetAppsRow struct {
	ID        int64      `json:"id"`
	Schema    AppPackage `json:"schema"`
	DateAdded int64      `json:"dateAdded"`
}

// GetApps Modified sqlc function as the JSON column needs parsing to a custom struct
func (q *Queries) GetApps(ctx context.Context) ([]GetAppsRow, error) {
	rows, err := q.db.QueryContext(ctx, getApps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GetAppsRow
	for rows.Next() {
		var i GetAppsRow
		var packageString string
		if err := rows.Scan(&i.ID, &packageString, &i.DateAdded); err != nil {
			return nil, err
		}

		err := json.Unmarshal([]byte(packageString), &i.Schema)
		if err != nil {
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
