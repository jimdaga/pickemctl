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

	uidrows, err := db.Query("SELECT DISTINCT(uid) FROM public.pickem_api_gamepicks WHERE gameseason IS NOT NULL")
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
			"WHERE uid = $1 AND pick_correct = true AND gameseason IS NOT NULL", uid).Scan(&correctPicksTotal)
		if err != nil {
			log.Printf("Error getting total correct picks for UID %s: %v", uid, err)
			continue
		}

		err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND gameseason IS NOT NULL", uid).Scan(&totalPicksTotal)
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
			"WHERE uid = $1 AND pick_correct = true AND gameseason = $2 AND gameseason IS NOT NULL", uid, currentSeason).Scan(&correctPicksSeason)
		if err != nil {
			log.Printf("Error getting season correct picks for UID %s: %v", uid, err)
			// Continue with just total stats
		} else {
			err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
				"WHERE uid = $1 AND gameseason = $2 AND gameseason IS NOT NULL", uid, currentSeason).Scan(&totalPicksSeason)
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
	uidrows, err := db.Query("SELECT DISTINCT(uid) FROM public.pickem_api_gamepicks WHERE gameseason IS NOT NULL")
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
			"COALESCE(SUM("+
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
			"), 0) AS \"total_wins\""+
			"FROM \"pickem_api_userseasonpoints\""+
			"WHERE \"userID\" = $1 AND \"gameseason\" IS NOT NULL "+
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
			"COALESCE(SUM("+
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
			"), 0) AS \"season_wins\""+
			"FROM \"pickem_api_userseasonpoints\""+
			"WHERE \"userID\" = $1 AND \"gameseason\" = $2 AND \"gameseason\" IS NOT NULL", uid, currentSeason).Scan(&weeksWonSeason)

		if err != nil {
			log.Printf("Error getting season weeks won for UID %s: %v", uid, err)
			weeksWonSeason = 0
		}

		// Set season weeks won
		stats.WeeksWonSeason = dbUtil.IntPtr(weeksWonSeason)

		// Calculate seasons won (year_winner = true count)
		var seasonsWon int
		err = db.QueryRow("SELECT COUNT(*) FROM \"pickem_api_userseasonpoints\" WHERE \"userID\" = $1 AND \"year_winner\" = true AND \"gameseason\" IS NOT NULL", uid).Scan(&seasonsWon)
		if err != nil {
			if err == sql.ErrNoRows {
				seasonsWon = 0
				log.Printf("No season winner records found for UID %s, setting seasons won to 0", uid)
			} else {
				log.Printf("Error getting seasons won for UID %s: %v", uid, err)
				seasonsWon = 0 // Default to 0 on error
			}
		}

		// Set seasons won
		stats.SeasonsWon = dbUtil.IntPtr(seasonsWon)

		// Calculate missed picks - season
		var missedPicksSeason int

		// Count scored games for current season that the user did NOT pick
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM "pickem_api_gamesandscores" gs
			WHERE gs."gameseason" = $1 
			AND gs."gameScored" = true
			AND gs."gameseason" IS NOT NULL
			AND NOT EXISTS (
				SELECT 1 FROM "pickem_api_gamepicks" gp 
				WHERE gp."pick_game_id" = gs."id" 
				AND gp."userID" = $2
				AND gp."gameseason" IS NOT NULL
			)`, currentSeason, uid).Scan(&missedPicksSeason)

		if err != nil {
			log.Printf("Error getting missed picks for season %s, UID %s: %v", currentSeason, uid, err)
			missedPicksSeason = 0
		}
		stats.MissedPicksSeason = dbUtil.IntPtr(missedPicksSeason)

		// Calculate missed picks - total (all time)
		var missedPicksTotal int

		// Count scored games across all seasons that the user did NOT pick
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM "pickem_api_gamesandscores" gs
			WHERE gs."gameScored" = true
			AND gs."gameseason" IS NOT NULL
			AND NOT EXISTS (
				SELECT 1 FROM "pickem_api_gamepicks" gp 
				WHERE gp."pick_game_id" = gs."id" 
				AND gp."uid" = $1
				AND gp."gameseason" IS NOT NULL
			)`, uid).Scan(&missedPicksTotal)

		if err != nil {
			log.Printf("Error getting total missed picks for UID %s: %v", uid, err)
			missedPicksTotal = 0
		}
		stats.MissedPicksTotal = dbUtil.IntPtr(missedPicksTotal)

		// Calculate perfect weeks - season
		var perfectWeeksSeason int
		perfectWeeksQuery := `
			SELECT COUNT(DISTINCT gs.gameweek) 
			FROM pickem_api_gamesandscores gs
			WHERE gs.gameseason = $1 
			AND gs.gamescored = true
			AND gs.gameseason IS NOT NULL
			AND (
				-- Count of scored games in this week
				SELECT COUNT(*) FROM pickem_api_gamesandscores gs2 
				WHERE gs2.gameweek = gs.gameweek 
				AND gs2.gameseason = gs.gameseason 
				AND gs2.gamescored = true
				AND gs2.gameseason IS NOT NULL
			) = (
				-- Count of correct picks by user in this week
				SELECT COUNT(*) FROM pickem_api_gamepicks gp 
				WHERE gp.gameweek = gs.gameweek 
				AND gp.gameseason = gs.gameseason 
				AND gp."uid" = $2 
				AND gp.pick_correct = true
				AND gp.gameseason IS NOT NULL
			)
			AND (
				-- Ensure user made picks for ALL scored games (no missed picks)
				SELECT COUNT(*) FROM pickem_api_gamesandscores gs3
				WHERE gs3.gameweek = gs.gameweek 
				AND gs3.gameseason = gs.gameseason 
				AND gs3.gamescored = true
				AND gs3.gameseason IS NOT NULL
			) = (
				-- Count of total picks by user in this week
				SELECT COUNT(*) FROM pickem_api_gamepicks gp2 
				WHERE gp2.gameweek = gs.gameweek 
				AND gp2.gameseason = gs.gameseason 
				AND gp2."uid" = $2
				AND gp2.gameseason IS NOT NULL
			)`

		err = db.QueryRow(perfectWeeksQuery, currentSeason, uid).Scan(&perfectWeeksSeason)
		if err != nil {
			log.Printf("Error getting perfect weeks for season %s, UID %s: %v", currentSeason, uid, err)
			perfectWeeksSeason = 0
		}
		stats.PerfectWeeksSeason = dbUtil.IntPtr(perfectWeeksSeason)

		// Calculate perfect weeks - total (all time)
		var perfectWeeksTotal int
		perfectWeeksTotalQuery := `
			SELECT COUNT(DISTINCT gs.gameseason || '-' || gs.gameweek) 
			FROM pickem_api_gamesandscores gs
			WHERE gs.gamescored = true
			AND gs.gameseason IS NOT NULL
			AND (
				-- Count of scored games in this week/season
				SELECT COUNT(*) FROM pickem_api_gamesandscores gs2 
				WHERE gs2.gameweek = gs.gameweek 
				AND gs2.gameseason = gs.gameseason 
				AND gs2.gamescored = true
				AND gs2.gameseason IS NOT NULL
			) = (
				-- Count of correct picks by user in this week/season
				SELECT COUNT(*) FROM pickem_api_gamepicks gp 
				WHERE gp.gameweek = gs.gameweek 
				AND gp.gameseason = gs.gameseason 
				AND gp."userID" = $1 
				AND gp.pick_correct = true
				AND gp.gameseason IS NOT NULL
			)
			AND (
				-- Ensure user made picks for ALL scored games (no missed picks)
				SELECT COUNT(*) FROM pickem_api_gamesandscores gs3
				WHERE gs3.gameweek = gs.gameweek 
				AND gs3.gameseason = gs.gameseason 
				AND gs3.gamescored = true
				AND gs3.gameseason IS NOT NULL
			) = (
				-- Count of total picks by user in this week/season
				SELECT COUNT(*) FROM pickem_api_gamepicks gp2 
				WHERE gp2.gameweek = gs.gameweek 
				AND gp2.gameseason = gs.gameseason 
				AND gp2."userID" = $1
				AND gp2.gameseason IS NOT NULL
			)`

		err = db.QueryRow(perfectWeeksTotalQuery, uid).Scan(&perfectWeeksTotal)
		if err != nil {
			log.Printf("Error getting total perfect weeks for UID %s: %v", uid, err)
			perfectWeeksTotal = 0
		}
		stats.PerfectWeeksTotal = dbUtil.IntPtr(perfectWeeksTotal)

		// Upsert the user stats
		if err := dbUtil.UpsertUserStats(db, stats); err != nil {
			log.Printf("Error upserting user stats for UID %s: %v", uid, err)
		} else {
			log.Printf("✓ UID: %s | Weeks Won: %d/%d | Seasons Won: %d | Missed Picks: %d/%d | Perfect Weeks: %d/%d", uid, weeksWonSeason, weeksWonTotal, seasonsWon, missedPicksSeason, missedPicksTotal, perfectWeeksSeason, perfectWeeksTotal)
		}
	}
}
