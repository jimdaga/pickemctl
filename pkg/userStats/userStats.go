package userStats

import (
	"github.com/jimdaga/pickemcli/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// UserStats represents the main userStats command
var UserStats = &cobra.Command{
	Use:   "userStats",
	Short: "Generate user statistics and analytics",
	Long: `User Statistics Generation
			Generate various analytics based on user picks including:
			- Pick accuracy statistics  
			- Most and least picked teams
			- Weekly wins tracking`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.Connect()
		defer database.Close()

		// Run all user statistics operations
		RunPickStats(database)
		RunTopPicked(database) 
		RunLeastPicked(database)
	},
}

// AllStats runs all user statistics operations
func AllStats() *cobra.Command {
	return UserStats
}

func init() {
	// Set configuration defaults
	viper.SetDefault("app.season.current", "2425")
}
