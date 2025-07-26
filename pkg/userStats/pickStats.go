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

// PickStats represents the pickStats command
var PickStats = &cobra.Command{
	Use:   "pickStats",
	Short: "Generate pick analytics",
	Long: `Picks Data Generation
			Generate various analytics based on users picks`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.Connect()
		defer database.Close()

		RunPickStats(database)
	},
}

// RunPickStats executes the pick statistics analysis
func RunPickStats(db *sql.DB) {
	fmt.Println("..| Correct Picks by UID |..")
	CorrectPicksByUid(db)
	fmt.Println("..| Weeks Won by UID |..")
	WeeksWonByUid(db)
}

func CorrectPicksByUid(db *sql.DB) {
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

	// Process each user
	for _, uid := range uids {
		// Get user email
		userEmail, err := dbUtil.GetUserEmail(db, uid)
		if err != nil {
			log.Printf("Error getting email for UID %s: %v", uid, err)
			continue
		}

		// Create user stats object
		stats := dbUtil.NewUserStats(uid, userEmail)

		// Calculate ALL TIME stats
		var correctPicksTotal, totalPicksTotal int
		err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND pick_correct = true", uid).Scan(&correctPicksTotal)
		if err != nil {
			log.Printf("Error getting total correct picks for UID %s: %v", uid, err)
			continue
		}

		err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1", uid).Scan(&totalPicksTotal)
		if err != nil {
			log.Printf("Error getting total picks for UID %s: %v", uid, err)
			continue
		}

		var percentTotal int
		if totalPicksTotal > 0 {
			percentTotal = int(float64(correctPicksTotal) / float64(totalPicksTotal) * 100)
		}

		// Set all-time stats
		stats.CorrectPickTotalTotal = dbUtil.IntPtr(correctPicksTotal)
		stats.TotalPicksTotal = dbUtil.IntPtr(totalPicksTotal)
		stats.PickPercentTotal = dbUtil.IntPtr(percentTotal)

		// Calculate CURRENT SEASON stats
		var correctPicksSeason, totalPicksSeason int
		err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND pick_correct = true AND gameseason = $2", uid, currentSeason).Scan(&correctPicksSeason)
		if err != nil {
			log.Printf("Error getting season correct picks for UID %s: %v", uid, err)
			// Continue with just total stats
		} else {
			err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
				"WHERE uid = $1 AND gameseason = $2", uid, currentSeason).Scan(&totalPicksSeason)
			if err != nil {
				log.Printf("Error getting season picks for UID %s: %v", uid, err)
			} else {
				var percentSeason int
				if totalPicksSeason > 0 {
					percentSeason = int(float64(correctPicksSeason) / float64(totalPicksSeason) * 100)
				}

				// Set season stats
				stats.CorrectPickTotalSeason = dbUtil.IntPtr(correctPicksSeason)
				stats.TotalPicksSeason = dbUtil.IntPtr(totalPicksSeason)
				stats.PickPercentSeason = dbUtil.IntPtr(percentSeason)
			}
		}

		// Upsert the user stats
		if err := dbUtil.UpsertUserStats(db, stats); err != nil {
			log.Printf("Error upserting user stats for UID %s: %v", uid, err)
		} else {
			log.Printf("✓ UID: %s, Total: %d/%d (%d%%), Season: %d/%d (%d%%)",
				uid, correctPicksTotal, totalPicksTotal, percentTotal,
				correctPicksSeason, totalPicksSeason,
				func() int {
					if totalPicksSeason > 0 {
						return int(float64(correctPicksSeason) / float64(totalPicksSeason) * 100)
					}
					return 0
				}())
		}
	}
}

func WeeksWonByUid(db *sql.DB) {
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

	// Process each user
	for _, uid := range uids {
		// Get user email
		userEmail, err := dbUtil.GetUserEmail(db, uid)
		if err != nil {
			log.Printf("Error getting email for UID %s: %v", uid, err)
			continue
		}

		// Create user stats object
		stats := dbUtil.NewUserStats(uid, userEmail)

		// Calculate weeks won - all time
		var userID string
		var weeksWonTotal int
		err = db.QueryRow("SELECT \"userID\","+
			"SUM("+
			"CASE WHEN \"week_1_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_2_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_3_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_4_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_5_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_6_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_7_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_8_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_9_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_10_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_11_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_12_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_13_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_14_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_15_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_16_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_17_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_18_winner\" THEN 1 ELSE 0 END"+
			") AS \"total_wins\""+
			"FROM \"pickem_api_userseasonpoints\""+
			"WHERE \"userID\" = $1 "+
			"GROUP BY \"userID\"", uid).Scan(&userID, &weeksWonTotal)

		if err != nil {
			if err == sql.ErrNoRows {
				// User has no records yet, set to 0
				weeksWonTotal = 0
				log.Printf("No season points record found for UID %s, setting weeks won to 0", uid)
			} else {
				log.Printf("Error getting total weeks won for UID %s: %v", uid, err)
				continue
			}
		}

		// Set weeks won total
		stats.WeeksWonTotal = dbUtil.IntPtr(weeksWonTotal)

		// Calculate current season weeks won
		currentSeason := viper.GetString("app.season.current")
		var weeksWonSeason int
		err = db.QueryRow("SELECT "+
			"SUM("+
			"CASE WHEN \"week_1_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_2_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_3_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_4_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_5_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_6_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_7_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_8_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_9_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_10_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_11_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_12_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_13_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_14_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_15_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_16_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_17_winner\" THEN 1 ELSE 0 END +"+
			"CASE WHEN \"week_18_winner\" THEN 1 ELSE 0 END"+
			") AS \"season_wins\""+
			"FROM \"pickem_api_userseasonpoints\""+
			"WHERE \"userID\" = $1 AND \"gameseason\" = $2", uid, currentSeason).Scan(&weeksWonSeason)

		if err != nil {
			if err == sql.ErrNoRows {
				// User has no records for current season yet, set to 0
				weeksWonSeason = 0
				log.Printf("No current season record found for UID %s, setting season weeks won to 0", uid)
			} else {
				log.Printf("Error getting season weeks won for UID %s: %v", uid, err)
				// Continue with just total stats
			}
		}

		// Set season weeks won
		stats.WeeksWonSeason = dbUtil.IntPtr(weeksWonSeason)

		// Upsert the user stats
		if err := dbUtil.UpsertUserStats(db, stats); err != nil {
			log.Printf("Error upserting user stats for UID %s: %v", uid, err)
		} else {
			log.Printf("✓ UID: %s, Weeks Won - Total: %d, Season: %d", uid, weeksWonTotal, weeksWonSeason)
		}
	}
}
