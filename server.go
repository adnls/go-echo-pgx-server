package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return err
	}
	return nil
}

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgresql://root:password@localhost:5432/datacatalog")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer dbpool.Close()

	server := echo.New()

	server.Use(middleware.Logger())

	server.Validator = &CustomValidator{validator: validator.New(validator.WithRequiredStructEnabled())}

	sayHello := func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}

	server.GET("/", sayHello)

	apiv1 := server.Group("/api/v1")
	assets := apiv1.Group("/assets")

	type Asset struct {
		Type        string           `json:"type" validate:"required"`
		Source      string           `json:"source" validate:"required"`
		Identifiers []string         `json:"identifiers" validate:"required"`
		Doc         string           `json:"doc"`
		Components  []map[string]any `json:"components"`
		Properties  map[string]any   `json:"properties"`
	}

	sql := `
		insert into assets (type, source, identifiers, doc, components, properties) 
		values (@Type, @Source, @Identifiers, @Doc, @Components, @Properties)
		on conflict (type, source, identifiers)
		do update set 
		type = excluded.type,
		source = excluded.source,
		identifiers = excluded.identifiers,
		doc = excluded.doc,
		components = excluded.components,
		properties = excluded.properties,
		modified_at = default;`

	assets.POST("", func(c echo.Context) error {
		asset := new(Asset)
		if err = c.Bind(asset); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err = c.Validate(asset); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		conn, err := dbpool.Acquire(context.Background())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if err = pgx.BeginFunc(context.Background(), conn, func(tx pgx.Tx) error {
			_, err := tx.Exec(context.Background(), sql, pgx.NamedArgs{
				"Type":        asset.Type,
				"Source":      asset.Source,
				"Identifiers": asset.Identifiers,
				"Doc":         asset.Doc,
				"Components":  asset.Components,
				"Properties":  asset.Properties,
			})
			return err
		}); err != nil {
			echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return c.NoContent(http.StatusOK)
	}).Name = "Upsert single asset"

	if data, err := json.MarshalIndent(server.Routes(), "", "  "); err != nil {
		log.Fatal(err.Error())
	} else {
		if err := os.WriteFile("routes.json", data, 0644); err != nil {
			log.Fatal(err.Error())
		}
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Println("Cannot ping db")
	}

	log.Println("Ping db OK")

	if err := server.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err.Error())
	}
}
