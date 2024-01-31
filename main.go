package main

import (
    "fmt"
    "log"
    "net/http"
		"ChartingApp-Backend/internal/finance"
)

func main() {
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
