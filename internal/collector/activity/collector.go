package activity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okumujustine/postgresome/internal/metrics"
)

type ActivityCollector struct {
	pool *pgxpool.Pool
}

func NewActivityCollector(pool *pgxpool.Pool) *ActivityCollector {
	return &ActivityCollector{
		pool: pool,
	}
}

func (c *ActivityCollector) GetDatabaseActivity(ctx context.Context) (*metrics.DatabaseActivitySnapshot, error) {
	query := `
		SELECT
			datname,
			pid,
			usename,
			application_name,
			client_addr::text,
			state,
			query,
			wait_event_type,
			wait_event,
			backend_start,
			query_start,
			state_change
		FROM pg_stat_activity
		WHERE pid <> pg_backend_pid();
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]metrics.DatabaseActivity, 0)

	for rows.Next() {
		var (
			databaseName     *string
			processID        int
			userName         *string
			applicationName  *string
			clientAddress    *string
			state            *string
			queryText        *string
			waitEventType    *string
			waitEvent        *string
			backendStartedAt time.Time
			queryStartedAt   *time.Time
			stateChangedAt   *time.Time
		)

		err := rows.Scan(
			&databaseName,
			&processID,
			&userName,
			&applicationName,
			&clientAddress,
			&state,
			&queryText,
			&waitEventType,
			&waitEvent,
			&backendStartedAt,
			&queryStartedAt,
			&stateChangedAt,
		)
		if err != nil {
			return nil, err
		}

		activities = append(activities, metrics.DatabaseActivity{
			DatabaseName:     stringValue(databaseName),
			ProcessID:        processID,
			UserName:         stringValue(userName),
			ApplicationName:  stringValue(applicationName),
			ClientAddress:    stringValue(clientAddress),
			State:            stringValue(state),
			Query:            stringValue(queryText),
			WaitEventType:    stringValue(waitEventType),
			WaitEvent:        stringValue(waitEvent),
			BackendStartedAt: backendStartedAt,
			QueryStartedAt:   queryStartedAt,
			StateChangedAt:   stateChangedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &metrics.DatabaseActivitySnapshot{
		CollectedAt: time.Now(),
		Activities:  activities,
	}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
