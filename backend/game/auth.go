package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var ( 
	oAuthConf *oauth2.Config
	oAuthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo"
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("AuthMiddleware triggered") // Logging
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("Authorization header is missing") // Logging
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userID", claims["sub"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
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

		client := oAuthConf.Client(context.Background(), token)
		response, err := client.Get(oAuthGoogleUrlAPI)
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
        fmt.Printf("Raw Google User Info Response: %s\n", contents)

        var userInfo struct {
            ID    string `json:"id"`
            Email string `json:"email"`
        }
        if err := json.Unmarshal(contents, &userInfo); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
            return
        }
        fmt.Printf("Parsed User Info - ID: %s, Email: %s\n", userInfo.ID, userInfo.Email)

        // Validate essential user info from provider
        if userInfo.Email == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "email not provided by OAuth provider"})
            return
        }

        // Find or create user in DB
        var user User
        db.Where(User{Email: userInfo.Email}).FirstOrCreate(&user, User{
            Email:        userInfo.Email,
            AuthProvider: "google",
            CreatedAt:    time.Now(),
        })

		// Generate JWT
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":   user.ID,
			"email": user.Email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Redirect to frontend with token
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?token=%s", os.Getenv("FRONTEND_URL"), tokenString))
	}
}
