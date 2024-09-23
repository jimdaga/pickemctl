package pickStats

import (
	"database/sql"
	"fmt"
	"github.com/jimdaga/pickemcli/internal/db"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"log"
)

// pickStats represents the pickStats command
var PickStats = &cobra.Command{
	Use:   "pickStats",
	Short: "Generate pick analytics",
	Long: `Picks Data Generation
			Generate various analytics based on users picks`,
	Run: func(cmd *cobra.Command, args []string) {
		db := db.Connect()
		defer db.Close()

		// Find what team each player has picked the most
		// this is for all time
		// Include if there is a tie for most picked
		fmt.Println("..| Correct Pics by UID |..")
		CorrectPicksByUid(db)
		WeeksWonByUid(db)
	},
}

func CorrectPicksByUid(db *sql.DB) {

	uidrows, err := db.Query("SELECT DISTINCT(uid) FROM public.pickem_api_gamepicks")
	if err != nil {
		fmt.Println(err)
	}
	defer uidrows.Close()

	uids := make([]string, 0)
	for uidrows.Next() {
		var uid string
		if err := uidrows.Scan(&uid); err != nil {
			fmt.Println(err)
		}
		uids = append(uids, uid)
	}

	var correctPicks int
	var totalPicks int

	// Figure out how many correct picks each user has
	for _, uid := range uids {
		err := db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1 AND pick_correct = true", uid).Scan(&correctPicks)
		if err != nil {
			fmt.Println("Error getting correct picks")
			fmt.Println(err)
			return
		}

		// Figure out how picks the user has ever submitted
		err = db.QueryRow("SELECT count(*) FROM pickem_api_gamepicks "+
			"WHERE uid = $1", uid).Scan(&totalPicks)
		if err != nil {
			fmt.Println("Error getting total picks")
			fmt.Println(err)
			return
		}

		percent := float64(correctPicks) / float64(totalPicks) * 100

		/* TODO: Update a database table with this information
		* TODO: Write django model to store this information */
		log.Printf(" - UID: %s, Correct Picks: %d, Total Picks: %d Percent Correct: %f \n", uid, correctPicks, totalPicks, percent)
	}
}

func WeeksWonByUid(db *sql.DB) {

	uidrows, err := db.Query("SELECT DISTINCT(uid) FROM public.pickem_api_gamepicks")
	if err != nil {
		fmt.Println(err)
	}
	defer uidrows.Close()

	uids := make([]string, 0)
	for uidrows.Next() {
		var uid string
		if err := uidrows.Scan(&uid); err != nil {
			fmt.Println(err)
		}
		uids = append(uids, uid)
	}

	// Figure out how many weeks won each user has
	for _, uid := range uids {
		var userID string
		var weeksWon int
		err := db.QueryRow("SELECT \"userID\","+
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
			"GROUP BY \"userID\"", uid).Scan(&userID, &weeksWon)
		if err != nil {
			fmt.Println("Error getting total weeks won")
			fmt.Println(err)
			return
		}

		/* TODO: Update a database table with this information
		* TODO: Write django model to store this information */
		log.Printf(" - UID: %s, Weeks Won: %d \n", uid, weeksWon)
	}
}

func init() {

}
