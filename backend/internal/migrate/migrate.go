package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// Run applies pending SQL migration files from dir, in filename order,
// tracking applied migrations in the schema_migrations table.
func Run(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if _, err := pool.Exec(ctx, createSchemaMigrationsTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		applied, err := isApplied(ctx, pool, file)
		if err != nil {
			return err
		}

		if applied {
			fmt.Printf("skipping already applied migration: %s\n", file)
			continue
		}

		if err := applyMigration(ctx, pool, dir, file); err != nil {
			return err
		}

		fmt.Printf("applied migration: %s\n", file)
	}

	return nil
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		files = append(files, entry.Name())
	}

	sort.Strings(files)

	return files, nil
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var exists bool

	err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status for %s: %w", version, err)
	}

	return exists, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, dir, file string) error {
	contents, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return fmt.Errorf("failed to read migration %s: %w", file, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for %s: %w", file, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(contents)); err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", file, err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
		return fmt.Errorf("failed to record migration %s: %w", file, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", file, err)
	}

	return nil
}
