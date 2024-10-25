package leastPicked

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jimdaga/pickemcli/internal/db"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"log"
)

// leastPicked represents the leastPicked command
var LeastPicked = &cobra.Command{
	Use:   "leastPicked",
	Short: "Generate least pick analytics",
	Long: `least Picks Data Generation
			Generate various analytics based on users least picks`,
	Run: func(cmd *cobra.Command, args []string) {
		db := db.Connect()
		defer db.Close()

		// Find what team each player has picked the most
		// this is for all time
		// Include if there is a tie for most picked
		fmt.Println("..| Most Picked Team(s) by UID |..")
		LeastPickedByUid(db)
	},
}

func LeastPickedByUid(db *sql.DB) {

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
			"HAVING COUNT(*) = (SELECT MIN(c) FROM (SELECT COUNT(*) as c FROM pickem_api_gamepicks WHERE uid = $1 GROUP BY uid, pick) subquery) "+
			"ORDER BY count ASC", uid)

		if err != nil {
			fmt.Println(err)
			return
		}

		defer current.Close()

		for current.Next() {
			var uid string
			var pick string
			var count int
			var exists bool
			_ = current.Scan(&uid, &pick, &count)

			if err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM public.pickem_api_userstats WHERE \"userID\"=$1)", uid).Scan(&exists); err != nil {
				log.Printf("error checking if row exists: %v", err)
			}

			if exists {
				// Row exists, perform an UPDATE
				_, err = db.Exec("UPDATE public.pickem_api_userstats SET \"leastPickedTotal\"=$1 WHERE \"userID\"=$2", pick, uid)
				if err != nil {
					log.Printf("error updating least picked total: %v", err)
				}
			} else {
				// Row does not exist, perform an INSERT
				_, err = db.Exec("INSERT INTO public.pickem_api_userstats (\"id\", \"userID\", \"leastPickedTotal\") VALUES ($1, $2, $3)", uuid.New(), uid, pick)
				if err != nil {
					log.Printf("error inserting new row: %v", err)
				}
			}
		}
	}
}

func init() {

}
