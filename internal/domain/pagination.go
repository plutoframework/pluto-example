package domain

import (
	"time"

	"github.com/plutoframework/pluto-core/pagination"
)

const CursorTimeLayout = "2006-01-02T15:04:05.000000000Z"

func AuditKey(event AuditEvent) pagination.Key {
	return pagination.Key{
		Sort: event.OccurredAt.UTC().Format(CursorTimeLayout),
		ID:   event.ID,
	}
}

func AuditBoundary(request pagination.Request) (time.Time, error) {
	if err := request.Validate(); err != nil {
		return time.Time{}, err
	}

	if request.After.ID == "" {
		return time.Time{}, nil
	}

	value, err := time.Parse(CursorTimeLayout, request.After.Sort)
	if err != nil {
		return time.Time{}, pagination.ErrInvalid
	}

	return value, nil
}
