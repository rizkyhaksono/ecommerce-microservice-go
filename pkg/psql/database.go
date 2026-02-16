package psql

import (
	"fmt"
	"os"
	"strings"

	"ecommerce-microservice-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadDatabaseConfig() (DatabaseConfig, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	var missingVars []string
	if host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if port == "" {
		missingVars = append(missingVars, "DB_PORT")
	}
	if user == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if password == "" {
		missingVars = append(missingVars, "DB_PASSWORD")
	}
	if dbName == "" {
		missingVars = append(missingVars, "DB_NAME")
	}
	if sslMode == "" {
		missingVars = append(missingVars, "DB_SSLMODE")
	}

	if len(missingVars) > 0 {
		return DatabaseConfig{}, fmt.Errorf("missing required database environment variables: %s", strings.Join(missingVars, ", "))
	}

	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
	}, nil
}

func (c DatabaseConfig) GetDSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode +
		" TimeZone=UTC"
}

// ConnectDB creates a new GORM database connection
func ConnectDB(loggerInstance *logger.Logger) (*gorm.DB, error) {
	cfg, err := LoadDatabaseConfig()
	if err != nil {
		loggerInstance.Error("Failed to load database configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to load database configuration: %w", err)
	}

	gormZap := logger.NewGormLogger(loggerInstance.Log).
		LogMode(gormlogger.Warn)

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormZap,
	})
	if err != nil {
		loggerInstance.Error("Error connecting to the database", zap.Error(err))
		return nil, err
	}

	loggerInstance.Info("Database connection successful")
	return db, nil
}

// AutoMigrate runs GORM AutoMigrate for the given models
func AutoMigrate(db *gorm.DB, loggerInstance *logger.Logger, models ...interface{}) error {
	err := db.AutoMigrate(models...)
	if err != nil {
		loggerInstance.Error("Error migrating database entities", zap.Error(err))
		return err
	}
	loggerInstance.Info("Database entities migration completed successfully")
	return nil
}
