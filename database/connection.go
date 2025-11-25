// database/connection.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
)

type Connection struct {
	PostgresDB  *sql.DB
	MongoClient *mongo.Client
}

func NewConnection() *Connection {
	godotenv.Load() // Load .env

	// ========== PostgreSQL ==========
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSL := os.Getenv("DB_SSLMODE")

	// DSN PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, dbSSL,
	)

	log.Println("Postgres DSN:", dsn)

	// Connect PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed open DB:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Failed connect DB:", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	log.Println("✅ PostgreSQL Connected!")

	// ========== MongoDB ==========
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURI := os.Getenv("MONGO_URI")
		mongoDB := os.Getenv("MONGO_DB")
		mongoURL = mongoURI + "/" + mongoDB
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatal("❌ Mongo connection failed:", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Mongo ping failed:", err)
	}

	log.Println("✅ Mongo Connected!")

	return &Connection{
		PostgresDB:  db,
		MongoClient: client,
	}
}
