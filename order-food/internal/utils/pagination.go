package utils

import (
	"fmt"
	"strconv"

	"github.com/shyampundkar/kart-challenge-workspace/order-food/internal/models"
)

// BuildPaginationLinks creates HATEOAS links for pagination
func BuildPaginationLinks(page, totalPages int, basePath string, perPage int) []models.Link {
	links := []models.Link{
		{Href: fmt.Sprintf("%s?page=%d&perPage=%d", basePath, page, perPage), Rel: "self", Method: "GET"},
	}

	if page > 1 {
		links = append(links, models.Link{
			Href:   fmt.Sprintf("%s?page=1&perPage=%d", basePath, perPage),
			Rel:    "first",
			Method: "GET",
		})
		links = append(links, models.Link{
			Href:   fmt.Sprintf("%s?page=%d&perPage=%d", basePath, page-1, perPage),
			Rel:    "prev",
			Method: "GET",
		})
	}

	if page < totalPages {
		links = append(links, models.Link{
			Href:   fmt.Sprintf("%s?page=%d&perPage=%d", basePath, page+1, perPage),
			Rel:    "next",
			Method: "GET",
		})
		links = append(links, models.Link{
			Href:   fmt.Sprintf("%s?page=%d&perPage=%d", basePath, totalPages, perPage),
			Rel:    "last",
			Method: "GET",
		})
	}

	return links
}

// ParseInt parses a string to int with a default value
func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(s)
	if err != nil || result < 1 {
		return defaultValue
	}
	return result
}
