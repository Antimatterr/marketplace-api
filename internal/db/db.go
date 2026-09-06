package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Pool tuning. Postgres forks a backend process per connection, so opening one
	// is expensive (TCP handshake + TLS + auth roundtrips + process fork); the pool
	// keeps connections alive for reuse instead of rebuilding them per request.
	//
	// MaxOpen caps concurrent connections so we queue on our side rather than
	// exhausting the server's max_connections (default 100, shared by all clients).
	// Past the cap, callers block in db.Query until a connection frees up or their
	// context is cancelled.
	//
	// MaxIdle is set equal to MaxOpen: with the default of 2, a burst would open 25
	// connections and immediately destroy 23 on release, recreating the churn the
	// pool exists to avoid.
	//
	// ConnMaxLifetime is not a perf knob -- it forces recycling so we don't hold
	// connections killed by a server-side idle timeout or proxy, and so failover or
	// a DNS change to the DB host gets picked up.
	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//fail fast
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.ping: %w", err)
	}
	return db, nil

}
