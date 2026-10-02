package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDatabase starts a PostgreSQL container, applies migrations from the given folders, and returns the DB URL and a cleanup function.
func SetupTestDatabase(ctx context.Context, migrationFolders []string) (string, func(), error) {
	if testDB := os.Getenv("TEST_DB_URL"); testDB != "" {
		pool, err := pgxpool.New(ctx, testDB)
		if err == nil {
			for _, folder := range migrationFolders {
				_ = applyMigrations(ctx, pool, folder)
			}
			pool.Close()
			return testDB, func() {}, nil
		}
	}

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgis/postgis:15-3.3"), // We use postgis because ride queries require ST_X/ST_Y
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return "", nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	// Apply migrations
	for _, folder := range migrationFolders {
		err := applyMigrations(ctx, pool, folder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to apply migrations in %s: %w", folder, err)
		}
	}

	cleanup := func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			fmt.Printf("failed to terminate container: %v\n", err)
		}
	}

	return connString, cleanup, nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		// If folder doesn't exist, we skip or return error? We'll return error so the caller provides valid paths
		return err
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 7 && e.Name()[len(e.Name())-7:] == ".up.sql" {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, file := range upFiles {
		content, err := os.ReadFile(filepath.Join(folder, file))
		if err != nil {
			return err
		}
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("error executing %s: %w", file, err)
		}
	}
	return nil
}
