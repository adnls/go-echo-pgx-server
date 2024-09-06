package models

import "time"

type Asset struct {
	Type        string           `json:"type" validate:"required"`
	Source      string           `json:"source" validate:"required"`
	Identifiers []string         `json:"identifiers" validate:"required,gt=0"`
	Doc         string           `json:"doc"`
	Components  []map[string]any `json:"components"`
	Properties  map[string]any   `json:"properties"`
}

type AssetSearchResult struct {
	Id          int              `json:"id" db:"id"`
	Type        string           `json:"type" db:"type"`
	Source      string           `json:"source" db:"source"`
	Identifiers []string         `json:"identifiers" db:"identifiers"`
	Doc         string           `json:"doc" db:"doc"`
	Components  []map[string]any `json:"components" db:"components"`
	Properties  map[string]any   `json:"properties" db:"properties"`
	CreatedAt   time.Time        `json:"createdAt" db:"created_at"`
	ModifiedAt  time.Time        `json:"modifiedAt" db:"modified_at"`
	Rank        float32          `json:"rank" db:"rank"`
	Content     string           `json:"content" db:"content"`
	RowCount    int              `json:"rowCount" db:"row_count"`
}

type AssetSearchParams struct {
	Q string `query:"q" validate:"required"`
}

type AssetRepository interface {
	SaveOne(a Asset) error
	SaveMany(a []Asset) error
	Search(q string) ([]AssetSearchResult, error)
}
