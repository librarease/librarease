package database

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// implements server/Service interface
type service struct {
	db    *gorm.DB
	noti  *notificationHub
	cache *redis.Client
}

//go:embed migrations/notification.sql
var notificationSQL string

func New(gormDB *gorm.DB, noti *pgx.Conn, redis *redis.Client) (*service, error) {

	db, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	`)
	if err != nil {
		return nil, err
	}

	// migrate the schema
	err = gormDB.AutoMigrate(
		User{},
		AuthUser{},
		Library{},
		Staff{},
		Book{},
		Membership{},
		Subscription{},
		Returning{},
		Lost{},
		Borrowing{},
		Notification{},
		PushToken{},
		Watchlist{},
		Collection{},
		Asset{},
		CollectionBooks{},
		// CollectionFollowers{},
		Job{},
		Review{},
	)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(notificationSQL); err != nil {
		return nil, err
	}

	var notiHub *notificationHub
	if noti != nil {
		if _, err := noti.Exec(context.TODO(), "LISTEN \"new_notification\""); err != nil {
			return nil, err
		}
		notiHub = NewNotificationHub(noti)
	}

	return &service{db: gormDB, noti: notiHub, cache: redis}, nil
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
	fmt.Println("Disconnected from database")
	return db.Close()
}
