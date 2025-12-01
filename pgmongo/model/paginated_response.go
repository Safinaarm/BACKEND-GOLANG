// File: BACKEND-UAS/pgmongo/model/paginated_response.go
package model

type PaginatedResponse[T any] struct {
	Data       []T     `json:"data"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int64   `json:"total"`
	TotalPages int     `json:"total_pages"`
}