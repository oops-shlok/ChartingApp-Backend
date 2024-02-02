package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
  "github.com/joho/godotenv"
	"ChartingApp-Backend/internal/auth"
	"ChartingApp-Backend/internal/finance"
	"ChartingApp-Backend/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env file")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME not set in .env file")
	}

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("COLLECTION_NAME not set in .env file")
	}

	if err := database.InitMongoDB(mongoURI, dbName, collectionName); err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	m := http.NewServeMux()

	m.HandleFunc("/getHistory", finance.GetHistory)
	m.HandleFunc("/register", auth.CreateUserHandler)
	m.HandleFunc("/login", auth.AuthenticateHandler)

	addr := ":8080"
	srv := http.Server{
		Handler: m,
		Addr:    addr,
	}

	fmt.Println("Server is running at port", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}