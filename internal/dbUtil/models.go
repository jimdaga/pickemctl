package dbUtil

import (
	"github.com/google/uuid"
)

// UserStats represents the userStats model from Django
// This struct matches the fields in the Django model
type UserStats struct {
	ID                     uuid.UUID `db:"id"`
	UserEmail              string    `db:"userEmail"`
	UserID                 string    `db:"userID"`
	WeeksWonSeason         *int      `db:"weeksWonSeason"`
	WeeksWonTotal          *int      `db:"weeksWonTotal"`
	PickPercentSeason      *int      `db:"pickPercentSeason"`
	PickPercentTotal       *int      `db:"pickPercentTotal"`
	CorrectPickTotalSeason *int      `db:"correctPickTotalSeason"`
	CorrectPickTotalTotal  *int      `db:"correctPickTotalTotal"`
	TotalPicksSeason       *int      `db:"totalPicksSeason"`
	TotalPicksTotal        *int      `db:"totalPicksTotal"`
	MostPickedSeason       *string   `db:"mostPickedSeason"`
	MostPickedTotal        *string   `db:"mostPickedTotal"`
	LeastPickedSeason      *string   `db:"leastPickedSeason"`
	LeastPickedTotal       *string   `db:"leastPickedTotal"`
	SeasonsWon             *int      `db:"seasonsWon"`
	MissedPicksSeason      *int      `db:"missedPicksSeason"`
	MissedPicksTotal       *int      `db:"missedPicksTotal"`
	PerfectWeeksSeason     *int      `db:"perfectWeeksSeason"`
	PerfectWeeksTotal      *int      `db:"perfectWeeksTotal"`
}

// NewUserStats creates a new UserStats with a generated UUID
func NewUserStats(userID, userEmail string) *UserStats {
	return &UserStats{
		ID:        uuid.New(),
		UserID:    userID,
		UserEmail: userEmail,
	}
}
