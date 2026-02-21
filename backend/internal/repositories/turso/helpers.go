package turso

import (
	"database/sql"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// newID generates a new hex ObjectID string for use as a primary key.
func newID() string {
	return primitive.NewObjectID().Hex()
}

// timeToStr formats a time.Time as RFC3339 TEXT for storage.
func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

// timePtrToStr formats a *time.Time; returns "" for nil.
func timePtrToStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return timeToStr(*t)
}

// strToTime parses an RFC3339 string back to time.Time.
func strToTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

// strToTimePtr parses an RFC3339 string to *time.Time; returns nil for empty.
func strToTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t := strToTime(s)
	return &t
}

// oidFromHex converts a hex string to primitive.ObjectID. Returns zero for empty/invalid.
func oidFromHex(hex string) primitive.ObjectID {
	if hex == "" {
		return primitive.NilObjectID
	}
	oid, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID
	}
	return oid
}

// jsonMarshal marshals v to a JSON string, returning "[]" for nil slices.
func jsonMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// nullStr converts a sql.NullString to plain string.
func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// boolToInt converts a Go bool to 0/1 for SQLite.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts a SQLite 0/1 to Go bool.
func intToBool(i int) bool {
	return i != 0
}
