package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// GetDatabaseConfig returns database configuration from viper or defaults
func GetDatabaseConfig() DatabaseConfig {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "pickem")
	viper.SetDefault("database.sslmode", "disable")

	return DatabaseConfig{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		Database: viper.GetString("database.name"),
		SSLMode:  viper.GetString("database.sslmode"),
	}
}

// Connect establishes a connection to the PostgreSQL database using configuration
func Connect() *sql.DB {
	config := GetDatabaseConfig()

	var psqlInfo string
	if config.Password != "" {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Database, config.SSLMode)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	return db
}
