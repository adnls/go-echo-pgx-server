package main

import (
	"context"
	"datacatalog/server/controllers"
	"datacatalog/server/models"
	"datacatalog/server/repositories"
	"encoding/json"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	server := echo.New()

	server.Logger.SetLevel(log.DEBUG)

	logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		server.Logger.Fatal(err)
	}

	defer logFile.Close()

	server.Logger.SetOutput(io.MultiWriter(os.Stdout, logFile))

	server.HideBanner = true
	server.HidePort = true

	db, err := pgxpool.New(context.TODO(), "postgresql://root:password@localhost:5432/datacatalog")
	if err != nil {
		server.Logger.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(context.TODO()); err != nil {
		server.Logger.Warn("Cannot ping db")
	} else {
		server.Logger.Info("Ping db OK")
	}

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
