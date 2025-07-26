package daemon

import (
	"log"
	"time"

	"database/sql"
	"github.com/jimdaga/pickemcli/internal/db"
	"github.com/jimdaga/pickemcli/pkg/userStats"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	log.Printf("Running User Statistics Collection:")
	
	// Run all user statistics operations
	userStats.RunPickStats(db)
	log.Printf("\n")

	userStats.RunTopPicked(db)
	log.Printf("\n")

	userStats.RunLeastPicked(db)
	log.Printf("\n")
}

// Daemon starts the daemon process
func daemon() {
	log.Printf("Starting daemon\n")
	log.Printf("\n")

	seconds := viper.GetDuration("daemon.interval") * time.Second
	ticker := time.NewTicker(seconds)
	log.Printf("Daemon interval: %v\n", seconds)
	log.Printf("\n")

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
	// Set configuration defaults
	viper.SetDefault("daemon.interval", 30)
	viper.SetDefault("app.season.current", "2425")
}
