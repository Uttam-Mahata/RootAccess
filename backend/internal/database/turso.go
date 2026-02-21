package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// TursoDB holds the Turso (libSQL) connection via database/sql.
var TursoDB *sql.DB

// ConnectTurso opens a connection to a Turso database.
// url is the libsql:// or https:// connection URL.
// token is the auth token (may be empty for local dev).
func ConnectTurso(url, token string) {
	dsn := url
	if token != "" {
		dsn = fmt.Sprintf("%s?authToken=%s", url, token)
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		log.Fatalf("Failed to open Turso connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping Turso database: %v", err)
	}

	TursoDB = db
	log.Println("Connected to Turso (libSQL) successfully")
}
