package daemon

import (
	"log"
	"time"

	"database/sql"
	"github.com/jimdaga/pickemcli/internal/db"
	"github.com/jimdaga/pickemcli/pkg/leastPicked"
	"github.com/jimdaga/pickemcli/pkg/pickStats"
	"github.com/jimdaga/pickemcli/pkg/topPicked"
	"github.com/spf13/cobra"
)

// Daemon represents the daemon command
var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start a daemon process",
	Long: `Start a daemon process that runs in loop collecting 
			data to populate the family-pickem.com website`,
	Run: func(cmd *cobra.Command, args []string) {
		daemon()
	},
}

func collectData(db *sql.DB) {
	// Find what team each player has picked the MOST
	log.Printf("Most Picked Team by UID:")
	topPicked.TopPickedByUid(db)
	log.Printf("\n")

	// Find what team each player has picked the LEAST
	log.Printf("Least Picked Team by UID:")
	leastPicked.LeastPickedByUid(db)
	log.Printf("\n")

	// Find what team each player has picked the LEAST
	log.Printf("Correct Picks by UID:")
	pickStats.CorrectPicksByUid(db)
	log.Printf("\n")

	// Find what team each player has picked the LEAST
	log.Printf("Weeks Won by UID:")
	pickStats.WeeksWonByUid(db)
	log.Printf("\n")

}

// Daemon starts the daemon process
func daemon() {
	log.Printf("Starting daemon\n")
	log.Printf("\n")
	/* TODO: Make time configurable */
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})

	db := db.Connect()
	defer db.Close()

	// Run the data collect once before entering the loop:
	collectData(db)

	// Run the data collect every N seconds
	go func() {
		for {
			select {
			case <-ticker.C:
				collectData(db)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	// Block forever
	select {}
}

func init() {
	// TODO: Add flags for setting time
}
