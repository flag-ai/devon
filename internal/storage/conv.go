package storage

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// timeFromPg converts a pgtype.Timestamptz to a time.Time; zero value
// when the row's value is NULL.
func timeFromPg(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

// pgTimestamptz builds a valid pgtype.Timestamptz from a time.Time.
func pgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
