package persistence

import (
	"context"
	"encoding/json"
)

type GetAppsRow struct {
	getAppUnparsedRow
	Schema AppPackage
}

func (q *Queries) parseAppQuery(unparsedRow getAppUnparsedRow) (GetAppsRow, error) {
	row := GetAppsRow{getAppUnparsedRow: unparsedRow}
	if err := json.Unmarshal([]byte(unparsedRow.Schema.(string)), &row.Schema); err != nil {
		return row, err
	}
	return row, nil
}

// GetApp Retrieves a single app from the database
func (q *Queries) GetApp(ctx context.Context, id string) (GetAppsRow, error) {
	unparsed, err := q.getAppUnparsed(ctx, id)
	if err != nil {
		return GetAppsRow{}, err
	}
	return q.parseAppQuery(unparsed)
}

// GetApps Modified sqlc function as the JSON column needs parsing to a custom struct
func (q *Queries) GetApps(ctx context.Context) ([]GetAppsRow, error) {
	unparsedRows, err := q.getAppsUnparsed(ctx)
	if err != nil {
		return nil, err
	}

	var rows []GetAppsRow
	for _, unparsedRow := range unparsedRows {
		parsedRow, err := q.parseAppQuery(getAppUnparsedRow(unparsedRow))
		if err != nil {
			return nil, err
		}
		rows = append(rows, parsedRow)
	}
	return rows, nil
}

type AppWithCreds struct {
	getAppsWithCredsUnparsedRow
	Schema AppPackage
}

func (q *Queries) parseAppWithCredsQuery(unparsedRow getAppsWithCredsUnparsedRow) (AppWithCreds, error) {
	row := AppWithCreds{getAppsWithCredsUnparsedRow: unparsedRow}
	if err := json.Unmarshal([]byte(unparsedRow.Schema.(string)), &row.Schema); err != nil {
		return row, err
	}
	return row, nil
}

func (q *Queries) GetAppWithCreds(ctx context.Context, id string) (AppWithCreds, error) {
	unparsedRow, err := q.getAppWithCredsUnparsed(ctx, id)
	if err != nil {
		return AppWithCreds{}, err
	}
	return q.parseAppWithCredsQuery(getAppsWithCredsUnparsedRow(unparsedRow))
}

func (q *Queries) GetAppsWithCreds(ctx context.Context) ([]AppWithCreds, error) {
	unparsedRows, err := q.getAppsWithCredsUnparsed(ctx)
	if err != nil {
		return nil, err
	}

	var rows []AppWithCreds
	for _, unparsedRow := range unparsedRows {
		row, err := q.parseAppWithCredsQuery(unparsedRow)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}
