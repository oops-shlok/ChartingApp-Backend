// set url in web sockets (set headers too) and then pass it in market feel GET request
package finance

import (
	"encoding/json"
  "fmt"
  "net/http"
  "net/url"
	"time"
	"os"
	"os/signal"
	"log"

	"ChartingApp-Backend/internal/database"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"context"
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

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	accessToken, err := GetAccessTokenFromDB()
	url := "wss://api.upstox.com/v2/feed/market-data-feed"

    // Connect to the WebSocket server
    c, _, err := websocket.DefaultDialer.Dial(url, http.Header{
			"Authorization": []string{"Bearer " + accessToken},
		})

    if err != nil {
        log.Fatal("Error connecting to WebSocket:", err)
    }
    defer c.Close()

		done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("Received message: %s\n", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("{'guid':'someguid','method':'sub','data':{'mode':'full','instrumentKeys':['NSE_INDEX|Nifty Bank']}}"))
			if err != nil {
				log.Println("write:", err, t)
				return
			}
		case <-interrupt:
			fmt.Println("Interrupt received, closing connection...")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
		
}

func GetAccessTokenFromDB() (string, error) {
	db := database.GetDB()

	collection := db.Database("ChartKraft").Collection("Real Time")
	ctx := context.TODO()

	var tokenDoc bson.M
	err := collection.FindOne(ctx, bson.M{}).Decode(&tokenDoc)
	if err != nil {
		return "", err
	}

	accessToken := tokenDoc["token"].(string)
	return accessToken, nil
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