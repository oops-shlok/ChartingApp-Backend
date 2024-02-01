package main

import (
	"fmt"
	"log"
	"net/http"

	"ChartingApp-Backend/internal/auth"
	"ChartingApp-Backend/internal/database"
	"ChartingApp-Backend/internal/finance"
)

func main() {
	if err := database.InitMongoDB(); err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}



	authService := auth.NewAuthService("YMinzYNPlPw8dnljfXhSMsVwdVgpnXfI")
	authMiddleware := auth.NewAuthMiddleware(authService, nil)

	m := http.NewServeMux()

	m.Handle("/getHistory", authMiddleware.MiddlewareFunc(http.HandlerFunc(finance.GetHistory)))
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
