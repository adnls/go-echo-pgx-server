package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func main() {
	dbpool, err := pgxpool.New(context.Background(), "postgresql://root:password@localhost:5432/datacatalog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		panic(err)
	} else {
		fmt.Println("Ping db OK")
		fmt.Println(dbpool.Stat())
		fmt.Println(dbpool.Config())
	}

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/api/v1/assets", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
