package main

import (
    "fmt"
    "log"
    "net/http"
		"ChartingApp-Backend/internal/finance"
		"ChartingApp-Backend/internal/database"
)

func main() {
	if err := database.InitMongoDB(); err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	m := http.NewServeMux()

	addr := ":8080"

	m.HandleFunc("/getHistory", finance.GetHistory)
	srv := http.Server{
		Handler: m,
		Addr:    addr,
	}

	fmt.Println("server is running at port ", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
