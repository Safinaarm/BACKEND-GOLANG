package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
)

type Connection struct {
	PostgresDB *gorm.DB
	MongoClient *mongo.Client
}

func NewConnection() *Connection {
	godotenv.Load() // Load .env

	// Postgres URL dari .env (prioritas full URI, fallback build)
	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dbSSL := os.Getenv("DB_SSLMODE")
		pgURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)
	}

	// Connect Postgres
	db, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Postgres connection failed:", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Connected to Postgres (BACKEND-UAS1)!")

	// Mongo URL dari .env (prioritas full URI)
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = os.Getenv("MONGO_URI") + "/" + os.Getenv("MONGO_DB")
	}

	// Connect Mongo
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatal("Mongo connection failed:", err)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	// Fix: Ping di Client (test server Mongo)
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Mongo ping failed:", err)
	}
	log.Println("Connected to Mongo (BACKEND-UAS)!")

	return &Connection{
		PostgresDB:  db,
		MongoClient: client,
	}
}