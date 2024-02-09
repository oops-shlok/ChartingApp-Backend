package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"ChartingApp-Backend/internal/auth"
	"ChartingApp-Backend/internal/database"
	"ChartingApp-Backend/internal/data"
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

	corsHandler := cors.Default().Handler

	m.HandleFunc("/getHistory", finance.GetHistory)
	m.HandleFunc("/getRealTimeData", finance.LoginAuthorisation)
	m.HandleFunc("/getCode", finance.GetToken)
	m.HandleFunc("/ws", finance.HandleWebSocket)

	m.HandleFunc("/register", auth.CreateUserHandler)
	m.HandleFunc("/login", auth.AuthenticateHandler)

	m.HandleFunc("/oauth/login", auth.OAuthLoginHandler)
	m.HandleFunc("/oauth/callback", auth.OAuthCallbackHandler)

	handler := corsHandler(m)

	addr := ":8080"
	srv := http.Server{
		Handler: handler,
		Addr:    addr,
	}

	fmt.Println("Server is running at port", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
