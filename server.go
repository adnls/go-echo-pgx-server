package main

import (
	"context"
	"datacatalog/server/controllers"
	"datacatalog/server/models"
	"datacatalog/server/repositories"
	"encoding/json"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	switch v := i.(type) {
	case *[]models.Asset:
		if err := cv.validator.Var(v, "dive"); err != nil {
			return err
		}
	default:
		if err := cv.validator.Struct(v); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	db, err := pgxpool.New(context.TODO(), "postgresql://root:password@localhost:5432/datacatalog")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	if err := db.Ping(context.TODO()); err != nil {
		log.Println("Cannot ping db")
	} else {
		log.Println("Ping db OK")
	}

	server := echo.New()

	server.Validator = &CustomValidator{validator: validator.New(validator.WithRequiredStructEnabled())}

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	apiv1 := server.Group("/api/v1")

	assets := apiv1.Group("/assets")

	assetRepository := repositories.NewAssetRepository(db)
	assetHandler := controllers.NewAssetHandler(assetRepository)

	assets.POST("", assetHandler.HandleSaveOne).Name = "Upsert single asset"
	assets.POST("/bulk", assetHandler.HandleSaveMany).Name = "Upsert multiple assets"
	assets.GET("/search", assetHandler.HandlerSearch).Name = "Search assets"

	data, err := json.MarshalIndent(server.Routes(), "", "  ")
	if err != nil {
		server.Logger.Fatal(err)
	}
	if err := os.WriteFile("routes.json", data, 0644); err != nil {
		server.Logger.Fatal(err)
	}

	server.Logger.Fatal(server.Start(":8080"))
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

// server.HTTPErrorHandler = customHTTPErrorHandler
