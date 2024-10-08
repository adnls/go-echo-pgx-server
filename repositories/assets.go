package repositories

import (
	"context"
	"datacatalog/server/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssetRepository struct {
	Db *pgxpool.Pool
}

func NewAssetRepository(db *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{
		Db: db,
	}
}

const saveAssetSQL = `
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

const searchAssetSQL = "select * from websearch_assets(@Q)"

func (repo *AssetRepository) SaveOne(asset models.Asset) error {
	args := pgx.NamedArgs{
		"Type":        asset.Type,
		"Source":      asset.Source,
		"Identifiers": asset.Identifiers,
		"Doc":         asset.Doc,
		"Components":  asset.Components,
		"Properties":  asset.Properties,
	}
	execFunc := func(tx pgx.Tx) error {
		_, err := tx.Exec(context.TODO(), saveAssetSQL, args)
		return err
	}
	return pgx.BeginFunc(context.TODO(), repo.Db, execFunc)
}

func (repo *AssetRepository) SaveMany(assets []models.Asset) error {
	execFunc := func(tx pgx.Tx) error {
		batch := new(pgx.Batch)
		for _, asset := range assets {
			args := pgx.NamedArgs{
				"Type":        asset.Type,
				"Source":      asset.Source,
				"Identifiers": asset.Identifiers,
				"Doc":         asset.Doc,
				"Components":  asset.Components,
				"Properties":  asset.Properties,
			}
			batch.Queue(saveAssetSQL, args)
		}
		return tx.SendBatch(context.TODO(), batch).Close()
	}
	return pgx.BeginFunc(context.TODO(), repo.Db, execFunc)
}

func (repo *AssetRepository) Search(q string) ([]models.AssetSearchResult, error) {
	rows, err := repo.Db.Query(context.TODO(), searchAssetSQL, pgx.NamedArgs{"Q": q})
	if err != nil {
		return nil, err
	}
	assetSearchResults, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.AssetSearchResult])
	if err != nil {
		return nil, err
	}
	return assetSearchResults, nil
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
// results := dbpool.SendBatch(context.TODO(), batch)
// defer results.Close()
