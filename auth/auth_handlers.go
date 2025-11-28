package auth

import (
	"bot-assist/ent"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var KeyVal = make(map[string]any)

/*
Authorized ORIGINS
https://localhost:8080
https://t.me

Authorized REDIRECTS
http://localhost:8080
https://t.me/own_drive_bot?start=
http://localhost:8080/callback/telegram
http://localhost:8080/callback/whatsapp
http://localhost:8080/callback/discord


*/

func Auth(state string, provider string) (string, error) {
	config := oauth2.Config{
		ClientID:     (KeyVal["web"]).(map[string]any)["client_id"].(string),
		ClientSecret: (KeyVal["web"]).(map[string]any)["client_secret"].(string),
		RedirectURL:  "http://localhost:8080/callback/",
		Scopes:       []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	combinedState := fmt.Sprintf("%s:%s", state, provider)
	url := config.AuthCodeURL(combinedState, oauth2.SetAuthURLParam("provider", provider), oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return url, nil
}

type GoogleResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int    `json:"expires_in"`
	IdToken               string `json:"id_token"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
	TokenType             string `json:"token_type"`
}

func AuthCallBack(c *gin.Context) {
	ctx := context.Background()
	Combinedstate := c.Query("state")
	sliceState := strings.Split(Combinedstate, ":")
	provider := sliceState[1]

	code := c.Query("code")
	grantType := "authorization_code"
	ClientID := (KeyVal["web"]).(map[string]any)["client_id"].(string)
	ClientSecret := (KeyVal["web"]).(map[string]any)["client_secret"].(string)
	body := url.Values{
		"code":          {code},
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"redirect_uri":  {"http://localhost:8080/callback/"},
		"grant_type":    {grantType},
	}
	request, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(body.Encode()))
	if err != nil {
		log.Default().Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Default().Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Default().Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
	}()
	if response.StatusCode != http.StatusOK {
		log.Default().Println("Error:", response.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	var tokenResponse GoogleResponse
	if err := json.NewDecoder(response.Body).Decode(&tokenResponse); err != nil {
		log.Default().Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	userCreate := ent.Client.User.Create()
	verify, err := jwt.Parse(tokenResponse.IdToken, func(jwt *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	email := verify.Claims.(jwt.MapClaims)["email"]
	name := verify.Claims.(jwt.MapClaims)["name"]
	googleDriveId := fmt.Sprintf("%v", verify.Claims.(jwt.MapClaims)["sub"])
	accessToken := tokenResponse.AccessToken
	refreshToken := tokenResponse.RefreshToken
	userCreate.SetGoogleDriveID(googleDriveId).SetEmail(email.(string)).SetName(name.(string)).SetAccessToken(accessToken).SetRefreshToken(refreshToken).SetTokenType(tokenResponse.TokenType)
	newUser, err := userCreate.Save(ctx)

	if err != nil {
		log.Default().Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	log.Default().Println("User created successfully: ", newUser.ID, newUser.Name)

	switch provider {
	case "telegram":
		http.Redirect(c.Writer, c.Request, "https://t.me/own_drive_bot?start="+Combinedstate, http.StatusFound)
		return
	case "whatsapp":
		http.Redirect(c.Writer, c.Request, "https://localhost:8080/callback/whatsapp", http.StatusFound)
		return
	case "discord":
		http.Redirect(c.Writer, c.Request, "https://localhost:8080/callback/discord", http.StatusFound)
		return
	}

	http.Redirect(c.Writer, c.Request, "https://localhost:8080/callback/telegram", http.StatusFound)
}
