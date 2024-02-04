package finance

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func GetHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	queries := r.URL.Query()

	instrumentKey := queries.Get("instrument_key")
	fromDate := "2018-01-01"
	toDate := time.Now().Format("2006-01-02")
	interval := queries.Get("interval")

	if instrumentKey == "" || fromDate == "" || toDate == "" {
		http.Error(w, "Invalid Params", http.StatusBadRequest)
		return
	}

	fromTimestamp, err := convertToUnixTimestamp(fromDate)
	if err != nil {
		http.Error(w, "From date not in proper format", http.StatusBadRequest)
		return
	}

	toTimestamp, err := convertToUnixTimestamp(toDate)
	if err != nil {
		http.Error(w, "To date not in proper format", http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("https://api.upstox.com/v2/historical-candle/%s/%s/%s/%s", instrumentKey, interval, toDate, fromDate)

	params := map[string]string{
		"from_date": strconv.FormatInt(fromTimestamp, 10),
		"to_date":   strconv.FormatInt(toTimestamp, 10),
		"interval":  interval,
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
