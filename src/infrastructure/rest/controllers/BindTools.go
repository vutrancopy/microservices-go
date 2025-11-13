package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/gbrayhan/microservices-go/src/domain"
	"github.com/gin-gonic/gin"
)

func BindJSON(c *gin.Context, request any) error {
	buf := make([]byte, 5120)
	num, _ := c.Request.Body.Read(buf)
	reqBody := string(buf[0:num])
	c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(reqBody)))
	err := c.ShouldBindJSON(request)
	c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(reqBody)))
	return err
}

func BindJSONMap(c *gin.Context, request *map[string]any) error {
	buf := make([]byte, 5120)
	num, _ := c.Request.Body.Read(buf)
	reqBody := buf[0:num]
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	err := json.Unmarshal(reqBody, &request)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	return err
}

// GetSearchFilters parses search filters from query parameters
func GetSearchFilters(ctx *gin.Context) (domain.DataFilters, error) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if page < 1 {
		return domain.DataFilters{}, errors.New("invalid page number")
	}
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filters := domain.DataFilters{
		Page:     page,
		PageSize: pageSize,
	}

	return filters, nil
}

type MessageResponse struct {
	Message string `json:"message"`
}

type SortByDataRequest struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type FieldDateRangeDataRequest struct {
	Field     string `json:"field"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}
