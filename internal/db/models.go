// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type RiverJobState string

const (
	RiverJobStateAvailable RiverJobState = "available"
	RiverJobStateCancelled RiverJobState = "cancelled"
	RiverJobStateCompleted RiverJobState = "completed"
	RiverJobStateDiscarded RiverJobState = "discarded"
	RiverJobStatePending   RiverJobState = "pending"
	RiverJobStateRetryable RiverJobState = "retryable"
	RiverJobStateRunning   RiverJobState = "running"
	RiverJobStateScheduled RiverJobState = "scheduled"
)

func (e *RiverJobState) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RiverJobState(s)
	case string:
		*e = RiverJobState(s)
	default:
		return fmt.Errorf("unsupported scan type for RiverJobState: %T", src)
	}
	return nil
}

type NullRiverJobState struct {
	RiverJobState RiverJobState
	Valid         bool // Valid is true if RiverJobState is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRiverJobState) Scan(value interface{}) error {
	if value == nil {
		ns.RiverJobState, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RiverJobState.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRiverJobState) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RiverJobState), nil
}

type Extension struct {
	ID                   int64
	VscExtensionID       string
	Name                 string
	DisplayName          string
	ShortDescription     pgtype.Text
	PublisherID          string
	PublisherName        string
	PublisherDisplayName string
	Installs             int32
	TrendingDaily        pgtype.Numeric
	TrendingWeekly       pgtype.Numeric
	TrendingMonthly      pgtype.Numeric
	WeightedRating       pgtype.Numeric
	PublishedAt          pgtype.Timestamp
	ReleasedAt           pgtype.Timestamp
	CreatedAt            pgtype.Timestamp
	UpdatedAt            pgtype.Timestamp
}

type Image struct {
	ID        int64
	ThemeID   int64
	Language  string
	Type      string
	Format    string
	Url       string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type RiverJob struct {
	ID          int64
	State       RiverJobState
	Attempt     int16
	MaxAttempts int16
	AttemptedAt pgtype.Timestamptz
	CreatedAt   pgtype.Timestamptz
	FinalizedAt pgtype.Timestamptz
	ScheduledAt pgtype.Timestamptz
	Priority    int16
	Args        []byte
	AttemptedBy []string
	Errors      [][]byte
	Kind        string
	Metadata    []byte
	Queue       string
	Tags        []string
}

type RiverLeader struct {
	ElectedAt pgtype.Timestamptz
	ExpiresAt pgtype.Timestamptz
	LeaderID  string
	Name      string
}

type RiverQueue struct {
	Name      string
	CreatedAt pgtype.Timestamptz
	Metadata  []byte
	PausedAt  pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type SchemaMigration struct {
	Version string
}

type Theme struct {
	ID                            int64
	ExtensionID                   int64
	Path                          string
	Name                          string
	DisplayName                   string
	EditorBackground              string
	EditorForeground              string
	ActivityBarBackground         string
	ActivityBarForeground         string
	ActivityBarInActiveForeground string
	ActivityBarBorder             *string
	ActivityBarActiveBorder       string
	ActivityBarActiveBackground   *string
	ActivityBarBadgeBackground    string
	ActivityBarBadgeForeground    string
	TabsContainerBackground       *string
	TabsContainerBorder           *string
	StatusBarBackground           *string
	StatusBarForeground           string
	StatusBarBorder               *string
	TabActiveBackground           *string
	TabInactiveBackground         *string
	TabActiveForeground           string
	TabBorder                     string
	TabActiveBorder               *string
	TabActiveBorderTop            *string
	TitleBarActiveBackground      string
	TitleBarActiveForeground      string
	TitleBarBorder                *string
	CreatedAt                     pgtype.Timestamp
	UpdatedAt                     pgtype.Timestamp
}