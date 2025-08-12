package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Message    string      `json:"message,omitempty"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, data interface{}, message ...string) {
	resp := Response{
		Success: true,
		Data:    data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	c.JSON(http.StatusOK, resp)
}

func CreatedResponse(c *gin.Context, data interface{}, message ...string) {
	resp := Response{
		Success: true,
		Data:    data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	c.JSON(http.StatusCreated, resp)
}

func ErrorResponse(c *gin.Context, statusCode int, err string) {
	resp := Response{
		Success: false,
		Error:   err,
	}
	c.JSON(statusCode, resp)
}

func BadRequestResponse(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusBadRequest, err)
}

func UnauthorizedResponse(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusUnauthorized, err)
}

func ForbiddenResponse(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusForbidden, err)
}

func NotFoundResponse(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusNotFound, err)
}

func InternalServerErrorResponse(c *gin.Context, err string) {
	ErrorResponse(c, http.StatusInternalServerError, err)
}

func PaginatedSuccessResponse(c *gin.Context, data interface{}, pagination Pagination, message ...string) {
	resp := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	c.JSON(http.StatusOK, resp)
}