package database

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var TursoDB *sql.DB

func ConnectTurso(url, token string) {
	dsn := url
	if token != "" {
		if strings.Contains(dsn, "?") {
			dsn += "&authToken=" + token
		} else {
			dsn += "?authToken=" + token
		}
	}

	// Use "libsql" driver instead of "turso"
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		log.Fatalf("Failed to open libsql connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping Turso: %v", err)
	}

	TursoDB = db
	log.Println("Connected to Turso successfully using libsql driver")
}
