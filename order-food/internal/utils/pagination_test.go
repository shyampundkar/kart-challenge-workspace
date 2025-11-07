package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt_ValidInput(t *testing.T) {
	result := ParseInt("42", 10)
	assert.Equal(t, 42, result)
}

func TestParseInt_EmptyString(t *testing.T) {
	result := ParseInt("", 10)
	assert.Equal(t, 10, result, "Should return default value for empty string")
}

func TestParseInt_InvalidInput(t *testing.T) {
	result := ParseInt("invalid", 10)
	assert.Equal(t, 10, result, "Should return default value for invalid input")
}

func TestParseInt_NegativeNumber(t *testing.T) {
	result := ParseInt("-5", 10)
	assert.Equal(t, 10, result, "Should return default value for negative number")
}

func TestParseInt_Zero(t *testing.T) {
	result := ParseInt("0", 10)
	assert.Equal(t, 10, result, "Should return default value for zero")
}

func TestParseInt_LargeNumber(t *testing.T) {
	result := ParseInt("1000", 10)
	assert.Equal(t, 1000, result)
}

func TestBuildPaginationLinks_FirstPage(t *testing.T) {
	links := BuildPaginationLinks(1, 5, "/api/product", 10)

	// Should have self, next, and last links
	assert.Len(t, links, 3)

	// Check self link
	assert.Equal(t, "/api/product?page=1&perPage=10", links[0].Href)
	assert.Equal(t, "self", links[0].Rel)

	// Check next link
	assert.Equal(t, "/api/product?page=2&perPage=10", links[1].Href)
	assert.Equal(t, "next", links[1].Rel)

	// Check last link
	assert.Equal(t, "/api/product?page=5&perPage=10", links[2].Href)
	assert.Equal(t, "last", links[2].Rel)
}

func TestBuildPaginationLinks_MiddlePage(t *testing.T) {
	links := BuildPaginationLinks(3, 5, "/api/product", 10)

	// Should have self, first, prev, next, and last links
	assert.Len(t, links, 5)

	// Check self link
	assert.Equal(t, "/api/product?page=3&perPage=10", links[0].Href)
	assert.Equal(t, "self", links[0].Rel)

	// Check first link
	assert.Equal(t, "/api/product?page=1&perPage=10", links[1].Href)
	assert.Equal(t, "first", links[1].Rel)

	// Check prev link
	assert.Equal(t, "/api/product?page=2&perPage=10", links[2].Href)
	assert.Equal(t, "prev", links[2].Rel)

	// Check next link
	assert.Equal(t, "/api/product?page=4&perPage=10", links[3].Href)
	assert.Equal(t, "next", links[3].Rel)

	// Check last link
	assert.Equal(t, "/api/product?page=5&perPage=10", links[4].Href)
	assert.Equal(t, "last", links[4].Rel)
}

func TestBuildPaginationLinks_LastPage(t *testing.T) {
	links := BuildPaginationLinks(5, 5, "/api/product", 10)

	// Should have self, first, and prev links
	assert.Len(t, links, 3)

	// Check self link
	assert.Equal(t, "/api/product?page=5&perPage=10", links[0].Href)
	assert.Equal(t, "self", links[0].Rel)

	// Check first link
	assert.Equal(t, "/api/product?page=1&perPage=10", links[1].Href)
	assert.Equal(t, "first", links[1].Rel)

	// Check prev link
	assert.Equal(t, "/api/product?page=4&perPage=10", links[2].Href)
	assert.Equal(t, "prev", links[2].Rel)
}

func TestBuildPaginationLinks_SinglePage(t *testing.T) {
	links := BuildPaginationLinks(1, 1, "/api/product", 10)

	// Should only have self link
	assert.Len(t, links, 1)

	// Check self link
	assert.Equal(t, "/api/product?page=1&perPage=10", links[0].Href)
	assert.Equal(t, "self", links[0].Rel)
}

func TestBuildPaginationLinks_DifferentBasePath(t *testing.T) {
	links := BuildPaginationLinks(1, 3, "/api/order", 20)

	// Check that base path is used correctly
	assert.Equal(t, "/api/order?page=1&perPage=20", links[0].Href)
	assert.Equal(t, "/api/order?page=2&perPage=20", links[1].Href)
	assert.Equal(t, "/api/order?page=3&perPage=20", links[2].Href)
}

func TestBuildPaginationLinks_AllLinksHaveMethod(t *testing.T) {
	links := BuildPaginationLinks(2, 3, "/api/product", 10)

	// Check that all links have GET method
	for _, link := range links {
		assert.Equal(t, "GET", link.Method)
	}
}

func TestBuildPaginationLinks_SecondPageOfTwo(t *testing.T) {
	links := BuildPaginationLinks(2, 2, "/api/product", 10)

	// Should have self, first, and prev links
	assert.Len(t, links, 3)

	assert.Equal(t, "self", links[0].Rel)
	assert.Equal(t, "first", links[1].Rel)
	assert.Equal(t, "prev", links[2].Rel)
}
