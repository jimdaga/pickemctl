package toppicks

import (
	"database/sql"
	"fmt"
	"github.com/jimdaga/pickemcli/internal/db"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// TopPicks represents the toppicks command
var TopPicks = &cobra.Command{
	Use:   "toppicks",
	Short: "Generate top pick analytics",
	Long: `Top Picks Data Generation
			Generate various analytics based on users top picks`,
	Run: func(cmd *cobra.Command, args []string) {
		db := db.Connect()
		defer db.Close()

		// Find what team each player has picked the most
		// this is for all time
		// Include if there is a tie for most picked
		fmt.Println("..| Most Picked Team(s) by UID |..")
		MostPicked(db)
	},
}

func MostPicked(db *sql.DB) {

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

	for _, uid := range uids {
		current, err := db.Query("SELECT uid, pick, COUNT(*) as count "+
			"FROM pickem_api_gamepicks "+
			"WHERE uid = $1 "+
			"GROUP BY uid, pick "+
			"HAVING COUNT(*) = (SELECT MAX(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 GROUP BY uid, pick) subquery) "+
			"ORDER BY count DESC", uid)

		if err != nil {
			fmt.Println(err)
			return
		}

		defer current.Close()

		for current.Next() {
			var uid string
			var pick string
			var count int
			err := current.Scan(&uid, &pick, &count)
			if err != nil {
				fmt.Println(err)
				continue
			}
			/* TODO: Update a database table with this information
			 * TODO: Write django model to store this information */
			fmt.Printf("UID: %s, Pick: %s, Count: %d\n", uid, pick, count)
		}
	}
}

func init() {

}
