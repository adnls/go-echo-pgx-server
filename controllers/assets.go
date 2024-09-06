package controllers

import (
	"datacatalog/server/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AssetHandler struct {
	repo models.AssetRepository
}

// NewBaseHandler returns a new BaseHandler
func NewAssetHandler(repo models.AssetRepository) *AssetHandler {
	return &AssetHandler{
		repo: repo,
	}
}

// HelloWorld returns Hello, World
func (hdlr *AssetHandler) HandleSaveOne(ctx echo.Context) error {
	el := new(models.Asset)
	if err := ctx.Bind(el); err != nil {
		return err
	}
	if err := ctx.Validate(el); err != nil {
		return err
	}
	if err := hdlr.repo.SaveOne(el); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}

func (hdlr *AssetHandler) HandleSaveMany(ctx echo.Context) error {
	el := new([]models.Asset)
	if err := ctx.Bind(el); err != nil {
		return err
	}
	if err := ctx.Validate(el); err != nil {
		return err
	}
	if err := hdlr.repo.SaveMany(el); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusOK)
}

func (hdlr *AssetHandler) HandlerSearch(ctx echo.Context) error {
	params := new(models.AssetSearchParams)
	if err := ctx.Bind(params); err != nil {
		return err
	}
	if err := ctx.Validate(params); err != nil {
		return err
	}
	assetSearchResults, err := hdlr.repo.Search(params.Q)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, assetSearchResults)
}
