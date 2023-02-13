package controllers

// Copyright (c) 2022 - Ceyhun Uzunoglu <ceyhunuzngl AT gmail dot com>

import (
	"context"
	"github.com/dmwm/CMSMonitoring/rucio-dataset-monitoring/models"
	mymongo "github.com/dmwm/CMSMonitoring/rucio-dataset-monitoring/mongo"
	"github.com/dmwm/CMSMonitoring/rucio-dataset-monitoring/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	// detailedDatasetsUniqueSortColumn required for pagination in order
	detailedDatasetsUniqueSortColumn = "_id"
)

// GetDetailedDatasets controller that returns datasets according to DataTable request json
func GetDetailedDatasets(collectionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// We need to provide models.DataTableCustomRequest to the controller initializer and use same type in casting
		ctx, cancel, req := InitializeCtxAndBindRequestBody(c, models.DataTableRequest{})
		defer cancel()

		c.JSON(http.StatusOK, getDetailedDatasetsResults(ctx, c, collectionName, req.(models.DataTableRequest)))
		return
	}
}

// getDetailedDatasetsResults get query results efficiently
func getDetailedDatasetsResults(ctx context.Context, c *gin.Context, collectionName string, req models.DataTableRequest) models.DatatableBaseResponse {
	collection := mymongo.GetCollection(collectionName)
	var detailedDatasets []models.DetailedDataset

	// Should use SearchBuilderRequest query
	searchQuery := mymongo.SearchQueryForSearchBuilderRequest(&req.SearchBuilderRequest)
	sortQuery := mymongo.SortQueryBuilder(&req, detailedDatasetsUniqueSortColumn)
	length := req.Length
	skip := req.Start

	cursor, err := mymongo.GetFindQueryResults(ctx, collection, searchQuery, sortQuery, skip, length)
	if err != nil {
		utils.ErrorResponse(c, "Find query failed", err, "")
	}
	if err = cursor.All(ctx, &detailedDatasets); err != nil {
		utils.ErrorResponse(c, "detailed datasets cursor failed", err, "")
	}

	totalRecCount := getFilteredCount(ctx, c, collection, searchQuery, req.Draw)
	filteredRecCount := totalRecCount
	return models.DatatableBaseResponse{
		Draw:            req.Draw,
		RecordsTotal:    totalRecCount,
		RecordsFiltered: filteredRecCount,
		Data:            detailedDatasets,
	}
}
