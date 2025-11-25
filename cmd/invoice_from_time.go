package cmd

import (
	"database/sql"
	"fmt"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/Andriiklymiuk/ung/internal/models"
)

// timeSessionGroup represents a group of time sessions for invoicing
type timeSessionGroup struct {
	ClientID     uint
	ClientName   string
	ContractID   *uint
	ContractName string
	ContractType string
	HourlyRate   *float64
	FixedPrice   *float64
	Currency     string
	TotalHours   float64
	Sessions     []models.TrackingSession
}

func getUnbilledTimeSessions() ([]timeSessionGroup, error) {
	// Query all billable sessions that haven't been invoiced yet
	// We check if notes contain "[Invoiced:" to see if already billed
	query := `
		SELECT
			ts.id, ts.client_id, ts.contract_id, ts.project_name, ts.start_time,
			ts.end_time, ts.duration, ts.hours, ts.notes,
			c.name as client_name,
			COALESCE(ct.name, '') as contract_name,
			COALESCE(ct.contract_type, 'hourly') as contract_type,
			ct.hourly_rate,
			ct.fixed_price,
			COALESCE(ct.currency, 'USD') as currency
		FROM tracking_sessions ts
		LEFT JOIN clients c ON ts.client_id = c.id
		LEFT JOIN contracts ct ON ts.contract_id = ct.id
		WHERE ts.billable = 1
		  AND ts.deleted_at IS NULL
		  AND (ts.notes NOT LIKE '%[Invoiced:%' OR ts.notes IS NULL)
		ORDER BY c.id, ct.id, ts.start_time
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group sessions by client/contract
	groupsMap := make(map[string]*timeSessionGroup)

	for rows.Next() {
		var session models.TrackingSession
		var clientName, contractName, contractType, currency string
		var clientID uint
		var contractID, duration sql.NullInt64
		var hourlyRate, fixedPrice sql.NullFloat64
		var hours sql.NullFloat64
		var endTime sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&session.ID,
			&clientID,
			&contractID,
			&session.ProjectName,
			&session.StartTime,
			&endTime,
			&duration,
			&hours,
			&notes,
			&clientName,
			&contractName,
			&contractType,
			&hourlyRate,
			&fixedPrice,
			&currency,
		)
		if err != nil {
			return nil, err
		}

		// Populate session fields
		session.ClientID = &clientID
		if contractID.Valid {
			cid := uint(contractID.Int64)
			session.ContractID = &cid
		}
		if endTime.Valid {
			session.EndTime = &endTime.Time
		}
		if duration.Valid {
			d := int(duration.Int64)
			session.Duration = &d
		}
		if hours.Valid {
			h := hours.Float64
			session.Hours = &h
		}
		if notes.Valid {
			session.Notes = notes.String
		}

		// Create group key
		groupKey := fmt.Sprintf("%d-%v", clientID, contractID.Int64)

		if _, exists := groupsMap[groupKey]; !exists {
			var rate *float64
			if hourlyRate.Valid {
				r := hourlyRate.Float64
				rate = &r
			}

			var fixed *float64
			if fixedPrice.Valid {
				f := fixedPrice.Float64
				fixed = &f
			}

			var cid *uint
			if contractID.Valid {
				c := uint(contractID.Int64)
				cid = &c
			}

			groupsMap[groupKey] = &timeSessionGroup{
				ClientID:     clientID,
				ClientName:   clientName,
				ContractID:   cid,
				ContractName: contractName,
				ContractType: contractType,
				HourlyRate:   rate,
				FixedPrice:   fixed,
				Currency:     currency,
				TotalHours:   0,
				Sessions:     []models.TrackingSession{},
			}
		}

		group := groupsMap[groupKey]
		group.Sessions = append(group.Sessions, session)
		if session.Hours != nil {
			group.TotalHours += *session.Hours
		}
	}

	// Convert map to slice
	var groups []timeSessionGroup
	for _, g := range groupsMap {
		if g.TotalHours > 0 {
			groups = append(groups, *g)
		}
	}

	return groups, nil
}

// getUnbilledTimeSessionsForClient gets unbilled sessions for a specific client
func getUnbilledTimeSessionsForClient(clientID uint) ([]timeSessionGroup, error) {
	groups, err := getUnbilledTimeSessions()
	if err != nil {
		return nil, err
	}

	// Filter to only this client
	var filtered []timeSessionGroup
	for _, g := range groups {
		if g.ClientID == clientID {
			filtered = append(filtered, g)
		}
	}

	return filtered, nil
}
