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

// TopPicked represents the topPicked command
var TopPicked = &cobra.Command{
	Use:   "topPicked",
	Short: "Generate top pick analytics",
	Long: `Top Picks Data Generation
			Generate various analytics based on users top picks`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.Connect()
		defer database.Close()

		RunTopPicked(database)
	},
}

// RunTopPicked executes the top picked teams analysis
func RunTopPicked(db *sql.DB) {
	fmt.Println("..| Most Picked Team(s) by UID |..")
	TopPickedByUid(db)
}

func TopPickedByUid(db *sql.DB) {
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

		// Find the most picked team for all time
		allTimeRows, err := db.Query("SELECT uid, pick, COUNT(*) as count "+
			"FROM pickem_api_gamepicks "+
			"WHERE uid = $1 "+
			"GROUP BY uid, pick "+
			"HAVING COUNT(*) = (SELECT MAX(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 GROUP BY uid, pick) subquery) "+
			"ORDER BY count DESC", uid)

		if err != nil {
			log.Printf("Error getting all-time most picked for UID %s: %v", uid, err)
			continue
		}

		// Collect all the most picked teams (handle ties)
		var mostPickedTeams []string
		var maxCount int
		for allTimeRows.Next() {
			var uidResult string
			var pick string
			var count int
			err := allTimeRows.Scan(&uidResult, &pick, &count)
			if err != nil {
				log.Printf("Error scanning all-time result for UID %s: %v", uid, err)
				continue
			}
			mostPickedTeams = append(mostPickedTeams, pick)
			maxCount = count
		}
		allTimeRows.Close()

		// Set the most picked team(s) for all time
		if len(mostPickedTeams) > 0 {
			// If multiple teams tied, join them with commas
			var mostPickedString string
			if len(mostPickedTeams) == 1 {
				mostPickedString = mostPickedTeams[0]
			} else {
				// Handle ties by joining team names
				mostPickedString = ""
				for i, team := range mostPickedTeams {
					if i > 0 {
						mostPickedString += ", "
					}
					mostPickedString += team
				}
			}
			stats.MostPickedTotal = dbUtil.StringPtr(mostPickedString)
		}

		// Find the most picked team for current season
		seasonRows, err := db.Query("SELECT uid, pick, COUNT(*) as count "+
			"FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND gameseason = $2 "+
			"GROUP BY uid, pick "+
			"HAVING COUNT(*) = (SELECT MAX(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 AND gameseason = $2 GROUP BY uid, pick) subquery) "+
			"ORDER BY count DESC", uid, currentSeason)

		if err != nil {
			log.Printf("Error getting season most picked for UID %s: %v", uid, err)
			// Continue with just all-time data
		} else {
			// Collect season most picked teams (handle ties)
			var seasonMostPickedTeams []string
			var seasonMaxCount int
			for seasonRows.Next() {
				var uidResult string
				var pick string
				var count int
				err := seasonRows.Scan(&uidResult, &pick, &count)
				if err != nil {
					log.Printf("Error scanning season result for UID %s: %v", uid, err)
					continue
				}
				seasonMostPickedTeams = append(seasonMostPickedTeams, pick)
				seasonMaxCount = count
			}
			seasonRows.Close()

			// Set the most picked team(s) for current season
			if len(seasonMostPickedTeams) > 0 {
				var seasonMostPickedString string
				if len(seasonMostPickedTeams) == 1 {
					seasonMostPickedString = seasonMostPickedTeams[0]
				} else {
					// Handle ties by joining team names
					seasonMostPickedString = ""
					for i, team := range seasonMostPickedTeams {
						if i > 0 {
							seasonMostPickedString += ", "
						}
						seasonMostPickedString += team
					}
				}
				stats.MostPickedSeason = dbUtil.StringPtr(seasonMostPickedString)
			}

			log.Printf("✓ UID: %s, Most Picked - Total: %s (%d picks), Season: %s (%d picks)",
				uid,
				func() string {
					if stats.MostPickedTotal != nil {
						return *stats.MostPickedTotal
					}
					return "none"
				}(),
				maxCount,
				func() string {
					if stats.MostPickedSeason != nil {
						return *stats.MostPickedSeason
					}
					return "none"
				}(),
				seasonMaxCount)
		}

		// Upsert the user stats
		if err := dbUtil.UpsertUserStats(db, stats); err != nil {
			log.Printf("Error upserting user stats for UID %s: %v", uid, err)
		}
	}
}
