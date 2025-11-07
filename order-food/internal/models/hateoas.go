package models

// Link represents a HATEOAS link
type Link struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// HATEOASResponse wraps any response with HATEOAS links
type HATEOASResponse struct {
	Data  interface{} `json:"data"`
	Links []Link      `json:"_links"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalPages int `json:"totalPages"`
	TotalItems int `json:"totalItems"`
}

// PaginatedResponse wraps paginated data with HATEOAS links
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Links      []Link         `json:"_links"`
}

// ProductWithLinks wraps a product with HATEOAS links
type ProductWithLinks struct {
	Product
	Links []Link `json:"_links"`
}

// OrderWithLinks wraps an order with HATEOAS links
type OrderWithLinks struct {
	Order
	Links []Link `json:"_links"`
}
