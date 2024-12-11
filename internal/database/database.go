package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// implements server/Service interface
type service struct {
	db *gorm.DB
}

var (
	database = os.Getenv("DB_DATABASE")
	password = os.Getenv("DB_PASSWORD")
	username = os.Getenv("DB_USER")
	port     = os.Getenv("DB_PORT")
	host     = os.Getenv("DB_HOST")
	// dbInstance *service
)

func New() *service {
	// Reuse Connection
	// if dbInstance != nil {
	// 	return dbInstance
	// }
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	// db, err := sql.Open("pgx", connStr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// dbInstance = &service{
	// 	db: db,
	// }
	// return dbInstance
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal(err)
	}

	db, err := gormDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	var maxOpenConnections int
	if m, err := strconv.Atoi(
		os.Getenv("DB_MAX_OPEN_CONNECTIONS")); err == nil {
		maxOpenConnections = m
	}
	db.SetMaxOpenConns(maxOpenConnections)

	// migrate the schema
	err = gormDB.AutoMigrate(
		User{},
		AuthUser{},
		Library{},
		Staff{},
		Book{},
		Membership{},
		Subscription{},
		Borrowing{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_book_id_returned_at_null
        ON borrowings (book_id)
        WHERE returned_at IS NULL
		AND deleted_at IS NULL;
    `)
	if err != nil {
		log.Fatal(err)
	}

	return &service{db: gormDB}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	db, _ := s.db.DB()

	// Ping the database
	err := db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	log.Printf("Disconnected from database: %s", database)
	return db.Close()
}
