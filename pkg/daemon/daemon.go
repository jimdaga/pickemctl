package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/jimdaga/pickemcli/internal/db"
	"github.com/jimdaga/pickemcli/pkg/toppicks"
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

// Daemon starts the daemon process
func daemon() {
	log.Printf("Starting daemon")
	/* TODO: Make time configurable */
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})

	db := db.Connect()
	defer db.Close()

	go func() {
		for {
			select {
			case <-ticker.C:
				// Find what team each player has picked the most
				fmt.Println("..| Most Picked Team(s) by UID |..")
				toppicks.MostPicked(db)
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
