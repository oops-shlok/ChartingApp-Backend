// set url in web sockets (set headers too) and then pass it in market feel GET request
package finance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"ChartingApp-Backend/internal/database"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func LoginAuthorisation(w http.ResponseWriter, r *http.Request) {
	clientID := "ccc72c99-4435-4c13-a227-10ed39f27634"
	redirectURI := "http://localhost:8080/getCode"
	if(clientID == "" || redirectURI == ""){
		http.Error(w, "Invalid Params", http.StatusBadRequest)
		return
	}
	authURL := fmt.Sprintf("https://api.upstox.com/v2/login/authorization/dialog?client_id=%s&redirect_uri=%s&response_type=code", clientID, redirectURI);
	http.Redirect(w,r,authURL, http.StatusFound);
}

func GetToken(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    code := r.Form.Get("code")
    if code == "" {
        http.Error(w, "Code parameter is required", http.StatusBadRequest)
        return
    }

		clientID := "ccc72c99-4435-4c13-a227-10ed39f27634"
		clientSecret := "cnz1ba3dhf"
		redirectURI := "http://localhost:8080/getCode"

    formData := url.Values{}
    formData.Set("code", code)
    formData.Set("client_id", clientID)
    formData.Set("client_secret", clientSecret)
    formData.Set("redirect_uri", redirectURI)
    formData.Set("grant_type", "authorization_code")

    resp, err := http.PostForm("https://api.upstox.com/v2/login/authorization/token", formData)
    if err != nil {
        http.Error(w, "Failed to send request to token endpoint", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        http.Error(w, "Failed to decode response body", http.StatusInternalServerError)
        return
    }

    accessToken, ok := response["access_token"].(string)
		if !ok {
			http.Error(w, "Access token not found in response", http.StatusInternalServerError)
			return
		}

		err = storeAccessTokenInDB(accessToken)
		if err != nil {
			http.Error(w, "Failed to store access token in database", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Access Token: %s", accessToken)
}

func storeAccessTokenInDB(accessToken string) error {
	db := database.GetDB()

	collection := db.Database("ChartKraft").Collection("Real Time")
	ctx := context.TODO()

	now := time.Now()

	expiry := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 30, 0, 0, now.Location())

	doc := bson.M{
		"token":   accessToken,
		"expiry":  expiry,
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	return nil
}

func GetAccessTokenFromDB(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	collection := db.Database("ChartKraft").Collection("Real Time")
	ctx := context.TODO()

	var tokenDoc bson.M
	err := collection.FindOne(ctx, bson.M{}).Decode(&tokenDoc)
	if err != nil {
		fmt.Println("Error Getting Access Token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessToken := tokenDoc["token"].(string)
	fmt.Fprint(w, string(accessToken))
}

func StoreOHLCDataInDB(data map[string]interface{}) error {
	db := database.GetDB()

	collection := db.Database("ChartKraft").Collection("OHLC Data")
	ctx := context.TODO()

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}


func GetLargeFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("file")
	file, err := os.Open("largeFiles/stocksLists.json")
	fmt.Println(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Decode JSON file content
	var data interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Convert data to JSON and send it to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
