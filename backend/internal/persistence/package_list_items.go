package persistence

import (
	"context"
	"unsafe"
)

type FullPackageListItem struct {
	PackageListItem
	Categories []string `json:"categories"`
}

func (q *Queries) InsertPackage(ctx context.Context, insertQuery FullPackageListItem) error {
	// Adds required categories to database
	for _, category := range insertQuery.Categories {
		err := q.addCategory(ctx, category)
		if err != nil {
			return err
		}
	}

	// Adds or updates package in database
	err := q.writePackageListItem(ctx, writePackageListItemParams(insertQuery.PackageListItem))
	if err != nil {
		return err
	}

	// Clear previous category definitions
	err = q.deletePackageCategoryDefinitions(ctx, insertQuery.ID)
	if err != nil {
		return err
	}

	// Adds package definitions in database
	for _, category := range insertQuery.Categories {
		err := q.addPackageCategoryDefinition(ctx, addPackageCategoryDefinitionParams{
			Category:  category,
			PackageID: insertQuery.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// handlePackageResult takes a list of rows from a query to the packages table joined to the categories table
// takes a getPackageListItemsRow slice, identical results from other queries should be cast to this before
// passing
func (q *Queries) handlePackageResult(result []getPackageListItemsRow) []FullPackageListItem {
	// creates objects with embedded slices for categories from results
	outIndex := -1
	currentId := ""
	var packages []FullPackageListItem
	for _, result := range result {
		if result.PackageListItem.ID != currentId {
			outIndex++
			currentId = result.PackageListItem.ID
			packages = append(packages, FullPackageListItem{
				PackageListItem: result.PackageListItem,
			})
		}
		if result.Category.Valid {
			packages[outIndex].Categories = append(packages[outIndex].Categories, result.Category.String)
		}
	}

	return packages
}

func (q *Queries) GetPackages(ctx context.Context) ([]FullPackageListItem, error) {
	// retrieves results rows from database
	queryResult, err := q.getPackageListItems(ctx)
	if err != nil {
		return nil, err
	}

	return q.handlePackageResult(queryResult), nil
}

func (q *Queries) GetPackage(ctx context.Context, packageID string) (FullPackageListItem, error) {
	queryResult, err := q.getPackageListItem(ctx, packageID)
	if err != nil {
		return FullPackageListItem{}, err
	}
	return q.handlePackageResult([]getPackageListItemsRow{getPackageListItemsRow(queryResult)})[0], nil
}

var _ = getPackageListItemsRow(searchPackageListItemsRow{})

func (q *Queries) SearchPackages(ctx context.Context, query string, category string, author string) ([]FullPackageListItem, error) {
	// Retrieves search results from DB
	queryResult, err := q.searchPackageListItems(ctx, searchPackageListItemsParams{
		ID:       query,
		Category: category,
		Author:   author,
	})
	if err != nil {
		return nil, err
	}

	// Casts results
	castedResults := *(*[]getPackageListItemsRow)(unsafe.Pointer(&queryResult))
	return q.handlePackageResult(castedResults), nil
}
