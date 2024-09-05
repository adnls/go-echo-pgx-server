package main

import (
	"context"
	"encoding/json"
	"fmt"
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

type Asset struct {
	Type        string           `json:"type" validate:"required"`
	Source      string           `json:"source" validate:"required"`
	Identifiers []string         `json:"identifiers" validate:"required,gt=0"`
	Doc         string           `json:"doc"`
	Components  []map[string]any `json:"components"`
	Properties  map[string]any   `json:"properties"`
}

func (cv *CustomValidator) Validate(i interface{}) error {
	switch v := i.(type) {
	case *Asset:
		if err := cv.validator.Struct(v); err != nil {
			return err
		}
	case *[]Asset:
		if err := cv.validator.Var(v, "dive"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported")
	}
	return nil
}

/*
	 type CustomError interface {
		error
		Send()
	}

	type ValidationError struct {
		Msg string
	}

	func (ve ValidationError) Error() string {
		return ve.Msg
	}

	func (ve ValidationError) Send(c echo.Context) {
		c.String(http.StatusBadRequest, ve.Error())
	}

	func customHTTPErrorHandler(err error, c echo.Context) {
		original, ok := err.(ValidationError)
		fmt.Printf("%T", original)
		if ok {
			original.Send(c)
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	}
*/
func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgresql://root:password@localhost:5432/datacatalog")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer dbpool.Close()

	server := echo.New()

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	server.Validator = &CustomValidator{validator: validator.New(validator.WithRequiredStructEnabled())}

	// server.HTTPErrorHandler = customHTTPErrorHandler

	sayHello := func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}

	server.GET("/", sayHello)

	apiv1 := server.Group("/api/v1")
	assets := apiv1.Group("/assets")

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
		if err := c.Bind(asset); err != nil {
			return err // ValidationError{Msg: err.Error()}
		}
		if err := c.Validate(asset); err != nil {
			return err // ValidationError{Msg: err.Error()}
		}

		namedArgs := pgx.NamedArgs{
			"Type":        asset.Type,
			"Source":      asset.Source,
			"Identifiers": asset.Identifiers,
			"Doc":         asset.Doc,
			"Components":  asset.Components,
			"Properties":  asset.Properties,
		}

		execFunc := func(tx pgx.Tx) error {
			_, err := tx.Exec(context.Background(), sql, namedArgs)
			return err
		}

		if err := pgx.BeginFunc(context.Background(), dbpool, execFunc); err != nil {
			return err
		}

		return c.String(http.StatusOK, "OK")
	}).Name = "Upsert single asset"

	assets.POST("/bulk", func(c echo.Context) error {
		assets := new([]Asset)
		if err := c.Bind(assets); err != nil {
			return err
		}
		if err := c.Validate(assets); err != nil {
			return err
		}

		batch := new(pgx.Batch)
		for _, asset := range *assets {
			args := pgx.NamedArgs{
				"Type":        asset.Type,
				"Source":      asset.Source,
				"Identifiers": asset.Identifiers,
				"Doc":         asset.Doc,
				"Components":  asset.Components,
				"Properties":  asset.Properties,
			}
			batch.Queue(sql, args)
		}

		execFunc := func(tx pgx.Tx) error {
			batch := new(pgx.Batch)
			for _, asset := range *assets {
				args := pgx.NamedArgs{
					"Type":        asset.Type,
					"Source":      asset.Source,
					"Identifiers": asset.Identifiers,
					"Doc":         asset.Doc,
					"Components":  asset.Components,
					"Properties":  asset.Properties,
				}
				batch.Queue(sql, args)
			}
			return tx.SendBatch(context.Background(), batch).Close()
		}

		if err := pgx.BeginFunc(context.Background(), dbpool, execFunc); err != nil {
			return err
		}

		// for i := 0; i < len(*assets); i++ {
		// 	_, err := res.Exec()
		// 	if err != nil {
		// 		// var pgErr *pgconn.PgError
		// 		// if errors.As(err, &pgErr) {
		// 		// 	log.Println(pgErr.Code)
		// 		// 	log.Println(pgErr.ConstraintName)
		// 		// 	return fmt.Errorf("Item at index problem", asset)
		// 		// }
		// 		return fmt.Errorf("item at position %d: %w", i, err)
		// 	}
		// }
		// return res.Close()
		// results := dbpool.SendBatch(context.Background(), batch)
		// defer results.Close()

		return c.String(http.StatusOK, "WIP")
	}).Name = "Upsert multiple assets"

	if data, err := json.MarshalIndent(server.Routes(), "", "  "); err != nil {
		log.Fatal(err.Error())
	} else {
		if err := os.WriteFile("routes.json", data, 0644); err != nil {
			log.Fatal(err.Error())
		}
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Println("Cannot ping db")
	} else {
		log.Println("Ping db OK")
	}

	if err := server.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err.Error())
	}
}
