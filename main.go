package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		next.ServeHTTP(w, r)
	})
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

	symbol := queries.Get("symbol")
	startDate := queries.Get("start_date")
	endDate := queries.Get("end_date")

	if symbol == "" || startDate == "" || endDate == "" {
		http.Error(w, "Invalid Params", http.StatusBadRequest)
		return
	}

	startTimestamp, err := convertToUnixTimestamp(startDate)
	if err != nil {
		http.Error(w, "Start date not in proper format", http.StatusBadRequest)
		return
	}

	endTimestamp, err := convertToUnixTimestamp(endDate)
	if err != nil {
		http.Error(w, "End date not in proper format", http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s", symbol)

	params := map[string]string{
		"period1":  strconv.FormatInt(startTimestamp, 10),
		"period2":  strconv.FormatInt(endTimestamp, 10),
		"interval": "1d",
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(body))
}

func convertToUnixTimestamp(dateString string) (int64, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, dateString)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func main() {
	m := http.NewServeMux()

	addr := ":8080"

	m.HandleFunc("/getHistory", getHistory)

	corsMux := corsHandler(m)

	srv := http.Server{
		Handler: corsMux,
		Addr:    addr,
	}

	fmt.Println("server is running at port ", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
