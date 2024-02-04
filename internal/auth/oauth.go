package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/google/uuid"

	"ChartingApp-Backend/internal/database"
)



var googleOauthConfig = &oauth2.Config{
	ClientID:  "392663706668-qlcvpuf7gsrk05m7qi9c9kn2vht4pmim.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL: "http://localhost:8080/oauth/callback",
	Endpoint:     google.Endpoint,
	Scopes:       []string{"profile", "email", "openid", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
}

func OAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	databaseName := os.Getenv("DB_NAME")
	if databaseName == "" {
		log.Fatal("DB_NAME not set in .env file")
	}

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("COLLECTION_NAME not set in .env file")
	}

	user, err := GetUserByUsername(userInfo["email"].(string), databaseName, collectionName)
	if err == nil {
		handleExistingUser(w, r, user)
		return
	}

	if err != ErrUserNotFound {
		http.Error(w, "Error checking user existence", http.StatusInternalServerError)
		return
	}

	handleNewUserRegistration(w, r, userInfo)
}

func getUserInfo(accessToken string) (map[string]interface{}, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func handleExistingUser(w http.ResponseWriter, r *http.Request, user User) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	jwtToken, err := generateJWTToken(user, secretKey)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: jwtToken,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleNewUserRegistration(w http.ResponseWriter, r *http.Request, userInfo map[string]interface{}) {
	databaseName := os.Getenv("DB_NAME")
	if databaseName == "" {
		log.Fatal("DB_NAME not set in .env file")
	}

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("COLLECTION_NAME not set in .env file")
	}

	newUser, err := createUserFromOAuth(userInfo)
	if err != nil {
		http.Error(w, "Failed to create/update user", http.StatusInternalServerError)
		return
	}

	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	jwtToken, err := generateJWTToken(newUser, secretKey)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: jwtToken,
		User:  newUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func createUserFromOAuth(userInfo map[string]interface{}) (User, error) {
	newUser := User{
		Username: userInfo["email"].(string),
	}

	randomPassword := generateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	newUser.Password = string(hashedPassword)

	databaseName := os.Getenv("DB_NAME")
	if databaseName == "" {
		log.Fatal("DB_NAME not set in .env file")
	}

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("COLLECTION_NAME not set in .env file")
	}

	collection := database.GetDB().Database(databaseName).Collection(collectionName)
	_, err = collection.InsertOne(context.TODO(), newUser)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func generateRandomPassword() string {
	passwordLength := 12
	randomPassword := uuid.New().String()
	return randomPassword[:passwordLength]
}
