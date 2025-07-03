package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var ( 
	oAuthConf *oauth2.Config
	oAuthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?alt=json"
)

func InitOAuth() {
	oAuthConf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func GenerateStateOauthCookie(c *gin.Context) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie("oauthstate", state, int(expiration.Unix()), "/", "localhost", false, true)
	return state
}

func GoogleLoginHandler(c *gin.Context) {
	state := GenerateStateOauthCookie(c)
	url := oAuthConf.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallbackHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := c.Cookie("oauthstate")
		if err != nil || c.Query("state") != state {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
			return
		}

		code := c.Query("code")
		token, err := oAuthConf.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "code exchange failed"})
			return
		}

		response, err := http.Get(oAuthGoogleUrlAPI + token.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
			return
		}
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
			return
		}

		var userInfo struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		json.Unmarshal(contents, &userInfo)

		// Find or create user in DB
		var user User
		db.Where(User{Email: userInfo.Email}).FirstOrCreate(&user, User{
			Email:        userInfo.Email,
			AuthProvider: "google",
			CreatedAt:    time.Now(),
		})

		// TODO: Generate session token and redirect to frontend
		c.JSON(http.StatusOK, gin.H{"message": "authentication successful", "user": user})
	}
}
