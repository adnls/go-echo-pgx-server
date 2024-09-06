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
	Id          int              `db:"id"`
	Type        string           `db:"type"`
	Source      string           `db:"source"`
	Identifiers []string         `db:"identifiers"`
	Doc         string           `db:"doc"`
	Components  []map[string]any `db:"components"`
	Properties  map[string]any   `db:"properties"`
	CreatedAt   time.Time        `db:"created_at"`
	ModifiedAt  time.Time        `db:"modified_at"`
	Rank        float32          `db:"rank"`
	Content     string           `db:"content"`
	RowCount    int              `db:"row_count"`
}

type AssetSearchParams struct {
	Q string `query:"q" validate:"required"`
}

type AssetRepository interface {
	SaveOne(a Asset) error
	SaveMany(a []Asset) error
	Search(q string) ([]AssetSearchResult, error)
}
