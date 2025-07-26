package dbUtil

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// UpsertUserStats performs an upsert operation (INSERT or UPDATE) for UserStats
// If a record exists for the given userID, it updates only the non-nil fields
// If no record exists, it creates a new one with the provided data
func UpsertUserStats(db *sql.DB, stats *UserStats) error {
	// Check if record exists
	exists, err := UserStatsExists(db, stats.UserID)
	if err != nil {
		return fmt.Errorf("error checking if user stats exists: %w", err)
	}

	if exists {
		return UpdateUserStats(db, stats)
	} else {
		return InsertUserStats(db, stats)
	}
}

// UserStatsExists checks if a UserStats record exists for the given userID
func UserStatsExists(db *sql.DB, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pickem_api_userstats WHERE "userID" = $1)`
	err := db.QueryRow(query, userID).Scan(&exists)
	return exists, err
}

// GetUserEmail fetches the email for a given user ID from the account_emailaddress table
// Returns a placeholder email if no record is found
func GetUserEmail(db *sql.DB, userID string) (string, error) {
	var email string
	query := `SELECT "email" FROM public.account_emailaddress WHERE "user_id" = $1`
	err := db.QueryRow(query, userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a placeholder email for users without email records
			placeholderEmail := fmt.Sprintf("user-%s@placeholder.local", userID)
			log.Printf("No email found for userID %s, using placeholder: %s", userID, placeholderEmail)
			return placeholderEmail, nil
		}
		return "", fmt.Errorf("error fetching user email for userID %s: %w", userID, err)
	}
	return email, nil
}

// InsertUserStats creates a new UserStats record
func InsertUserStats(db *sql.DB, stats *UserStats) error {
	query := `
		INSERT INTO pickem_api_userstats (
			"id", "userEmail", "userID", "weeksWonSeason", "weeksWonTotal",
			"pickPercentSeason", "pickPercentTotal", "correctPickTotalSeason",
			"correctPickTotalTotal", "totalPicksSeason", "totalPicksTotal",
			"mostPickedSeason", "mostPickedTotal", "leastPickedSeason",
			"leastPickedTotal", "seasonsWon"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := db.Exec(query,
		stats.ID, stats.UserEmail, stats.UserID, stats.WeeksWonSeason, stats.WeeksWonTotal,
		stats.PickPercentSeason, stats.PickPercentTotal, stats.CorrectPickTotalSeason,
		stats.CorrectPickTotalTotal, stats.TotalPicksSeason, stats.TotalPicksTotal,
		stats.MostPickedSeason, stats.MostPickedTotal, stats.LeastPickedSeason,
		stats.LeastPickedTotal, stats.SeasonsWon,
	)

	if err != nil {
		return fmt.Errorf("error inserting user stats: %w", err)
	}

	log.Printf("Inserted new UserStats record for userID: %s", stats.UserID)
	return nil
}

// UpdateUserStats updates an existing UserStats record with only the non-nil fields
func UpdateUserStats(db *sql.DB, stats *UserStats) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	// Build dynamic UPDATE query based on non-nil fields
	if stats.UserEmail != "" {
		setParts = append(setParts, fmt.Sprintf(`"userEmail" = $%d`, argIndex))
		args = append(args, stats.UserEmail)
		argIndex++
	}
	if stats.WeeksWonSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"weeksWonSeason" = $%d`, argIndex))
		args = append(args, *stats.WeeksWonSeason)
		argIndex++
	}
	if stats.WeeksWonTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"weeksWonTotal" = $%d`, argIndex))
		args = append(args, *stats.WeeksWonTotal)
		argIndex++
	}
	if stats.PickPercentSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"pickPercentSeason" = $%d`, argIndex))
		args = append(args, *stats.PickPercentSeason)
		argIndex++
	}
	if stats.PickPercentTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"pickPercentTotal" = $%d`, argIndex))
		args = append(args, *stats.PickPercentTotal)
		argIndex++
	}
	if stats.CorrectPickTotalSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"correctPickTotalSeason" = $%d`, argIndex))
		args = append(args, *stats.CorrectPickTotalSeason)
		argIndex++
	}
	if stats.CorrectPickTotalTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"correctPickTotalTotal" = $%d`, argIndex))
		args = append(args, *stats.CorrectPickTotalTotal)
		argIndex++
	}
	if stats.TotalPicksSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"totalPicksSeason" = $%d`, argIndex))
		args = append(args, *stats.TotalPicksSeason)
		argIndex++
	}
	if stats.TotalPicksTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"totalPicksTotal" = $%d`, argIndex))
		args = append(args, *stats.TotalPicksTotal)
		argIndex++
	}
	if stats.MostPickedSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"mostPickedSeason" = $%d`, argIndex))
		args = append(args, *stats.MostPickedSeason)
		argIndex++
	}
	if stats.MostPickedTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"mostPickedTotal" = $%d`, argIndex))
		args = append(args, *stats.MostPickedTotal)
		argIndex++
	}
	if stats.LeastPickedSeason != nil {
		setParts = append(setParts, fmt.Sprintf(`"leastPickedSeason" = $%d`, argIndex))
		args = append(args, *stats.LeastPickedSeason)
		argIndex++
	}
	if stats.LeastPickedTotal != nil {
		setParts = append(setParts, fmt.Sprintf(`"leastPickedTotal" = $%d`, argIndex))
		args = append(args, *stats.LeastPickedTotal)
		argIndex++
	}
	if stats.SeasonsWon != nil {
		setParts = append(setParts, fmt.Sprintf(`"seasonsWon" = $%d`, argIndex))
		args = append(args, *stats.SeasonsWon)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add userID as the last parameter for WHERE clause
	args = append(args, stats.UserID)

	query := fmt.Sprintf(`UPDATE pickem_api_userstats SET %s WHERE "userID" = $%d`,
		strings.Join(setParts, ", "), argIndex)

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error updating user stats: %w", err)
	}

	log.Printf("Updated UserStats record for userID: %s", stats.UserID)
	return nil
}

// Helper functions to create pointers for setting values
func IntPtr(i int) *int {
	return &i
}

func StringPtr(s string) *string {
	return &s
}
