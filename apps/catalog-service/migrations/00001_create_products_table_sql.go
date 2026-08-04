package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateProductsTableUpSql, downCreateProductsTableUpSql)
}

func upCreateProductsTableUpSql(ctx context.Context, tx *sql.Tx) error {
	query := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS categories (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS products (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		category_id UUID NOT NULL REFERENCES categories(id),
		name VARCHAR(255) NOT NULL,
		img_urls TEXT[] DEFAULT '{}',
		price_amount BIGINT NOT NULL,
		price_currency VARCHAR(3) DEFAULT 'IDR',
		stock INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	_,err := tx.ExecContext(ctx, query)
	return err
}

func downCreateProductsTableUpSql(ctx context.Context, tx *sql.Tx) error {
	query := `
	DROP TABLE IF EXISTS products;
	DROP TABLE IF EXISTS categories;
	`
	_, err := tx.ExecContext(ctx, query)
	return err
}
