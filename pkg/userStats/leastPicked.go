package userStats

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jimdaga/pickemcli/internal/db"
	"github.com/jimdaga/pickemcli/internal/dbUtil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LeastPicked represents the leastPicked command
var LeastPicked = &cobra.Command{
	Use:   "leastPicked",
	Short: "Generate least pick analytics",
	Long: `least Picks Data Generation
			Generate various analytics based on users least picks`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.Connect()
		defer database.Close()

		RunLeastPicked(database)
	},
}

// RunLeastPicked executes the least picked teams analysis
func RunLeastPicked(db *sql.DB) {
	fmt.Println("..| Least Picked Team(s) by UID |..")
	LeastPickedByUid(db)
}

func LeastPickedByUid(db *sql.DB) {
	currentSeason := viper.GetString("app.season.current")

	uidrows, err := db.Query("SELECT DISTINCT(uid) FROM public.pickem_api_gamepicks")
	if err != nil {
		log.Printf("Error getting distinct UIDs: %v", err)
		return
	}
	defer uidrows.Close()

	uids := make([]string, 0)
	for uidrows.Next() {
		var uid string
		if err := uidrows.Scan(&uid); err != nil {
			log.Printf("Error scanning UID: %v", err)
			continue
		}
		uids = append(uids, uid)
	}

	for _, uid := range uids {
		// Get user email
		userEmail, err := dbUtil.GetUserEmail(db, uid)
		if err != nil {
			log.Printf("Error getting email for UID %s: %v", uid, err)
			continue
		}

		// Create user stats object
		stats := dbUtil.NewUserStats(uid, userEmail)

		// Find the least picked team for all time
		allTimeRows, err := db.Query("SELECT uid, pick, COUNT(*) as count "+
			"FROM pickem_api_gamepicks "+
			"WHERE uid = $1 "+
			"GROUP BY uid, pick "+
			"HAVING COUNT(*) = (SELECT MIN(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 GROUP BY uid, pick) subquery) "+
			"ORDER BY count ASC", uid)
		if err != nil {
			log.Printf("Error getting all-time least picked for UID %s: %v", uid, err)
			continue
		}

		// Collect all the least picked teams (handle ties)
		var leastPickedTeams []string
		var minCount int
		for allTimeRows.Next() {
			var uidResult string
			var pick string
			var count int
			err := allTimeRows.Scan(&uidResult, &pick, &count)
			if err != nil {
				log.Printf("Error scanning all-time result for UID %s: %v", uid, err)
				continue
			}
			leastPickedTeams = append(leastPickedTeams, pick)
			minCount = count
		}
		allTimeRows.Close()

		// Set the least picked team(s) for all time
		if len(leastPickedTeams) > 0 {
			var leastPickedString string
			if len(leastPickedTeams) == 1 {
				leastPickedString = leastPickedTeams[0]
			} else {
				// Handle ties by joining team names
				leastPickedString = ""
				for i, team := range leastPickedTeams {
					if i > 0 {
						leastPickedString += ", "
					}
					leastPickedString += team
				}
			}
			stats.LeastPickedTotal = dbUtil.StringPtr(leastPickedString)
		}

		// Find the least picked team for current season
		seasonRows, err := db.Query("SELECT uid, pick, COUNT(*) as count "+
			"FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND gameseason = $2 "+
			"GROUP BY uid, pick "+
			"HAVING COUNT(*) = (SELECT MIN(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 AND gameseason = $2 GROUP BY uid, pick) subquery) "+
			"ORDER BY count ASC", uid, currentSeason)

		if err != nil {
			log.Printf("Error getting season least picked for UID %s: %v", uid, err)
			// Continue with just all-time data
		} else {
			// Collect season least picked teams (handle ties)
			var seasonLeastPickedTeams []string
			var seasonMinCount int
			for seasonRows.Next() {
				var uidResult string
				var pick string
				var count int
				err := seasonRows.Scan(&uidResult, &pick, &count)
				if err != nil {
					log.Printf("Error scanning season result for UID %s: %v", uid, err)
					continue
				}
				seasonLeastPickedTeams = append(seasonLeastPickedTeams, pick)
				seasonMinCount = count
			}
			seasonRows.Close()

			// Set the least picked team(s) for current season
			if len(seasonLeastPickedTeams) > 0 {
				var seasonLeastPickedString string
				if len(seasonLeastPickedTeams) == 1 {
					seasonLeastPickedString = seasonLeastPickedTeams[0]
				} else {
					// Handle ties by joining team names
					seasonLeastPickedString = ""
					for i, team := range seasonLeastPickedTeams {
						if i > 0 {
							seasonLeastPickedString += ", "
						}
						seasonLeastPickedString += team
					}
				}
				stats.LeastPickedSeason = dbUtil.StringPtr(seasonLeastPickedString)
			}

			log.Printf("✓ UID: %s, Least Picked - Total: %s (%d picks), Season: %s (%d picks)",
				uid,
				func() string {
					if stats.LeastPickedTotal != nil {
						return *stats.LeastPickedTotal
					}
					return "none"
				}(),
				minCount,
				func() string {
					if stats.LeastPickedSeason != nil {
						return *stats.LeastPickedSeason
					}
					return "none"
				}(),
				seasonMinCount)
		}

		// Upsert the user stats
		if err := dbUtil.UpsertUserStats(db, stats); err != nil {
			log.Printf("Error upserting user stats for UID %s: %v", uid, err)
		}
	}
}
