package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/pkg/commons"
)

const Userkey = "user"

func CreateSession(r *gin.Engine, apiPwd string) error {
	cookiePwd := []byte(apiPwd)
	if len(cookiePwd) == 0 {
		cookiePwd = make([]byte, 64)
		_, err := rand.Read(cookiePwd)
		if err != nil {
			return errors.Wrap(err, "Generating random key")
		}
		log.Debug().Msg("No password set so generated random key for cookie store")
	}
	store := sessions.NewCookieStore(cookiePwd)
	store.Options(sessions.Options{MaxAge: 86400, Path: "/"})
	r.Use(sessions.Sessions("torq_session", store))
	return nil
}

func RefreshCookieFile(cookiePath string) error {
	if cookiePath == "" {
		return nil
	}
	cookie := make([]byte, 64)
	_, err := rand.Read(cookie)
	if err != nil {
		return errors.Wrap(err, "Generating random key")
	}
	hexCookie := hex.EncodeToString(cookie)
	f, err := os.Create(cookiePath)
	if err != nil {
		return errors.Wrap(err, "Creating or truncating cookie file")
	}
	defer f.Close()
	_, err = f.WriteString(hexCookie)
	if err != nil {
		return errors.Wrap(err, "Writing to cookie file")
	}
	err = f.Sync()
	if err != nil {
		return errors.Wrap(err, "Flushing cookie contents to disk")
	}
	return nil
}

// TorqRequired checks the status of the torq service
func TorqRequired(c *gin.Context) {
	torqService := commons.GetCurrentTorqServiceState(commons.TorqService)
	if torqService.Status != commons.ServiceActive {
		c.AbortWithStatusJSON(http.StatusFailedDependency, gin.H{"error": "initializing"})
		return
	}
	c.Next()
}

// AuthRequired is a simple middleware to check the session
func AuthRequired(autoLogin bool) gin.HandlerFunc {
	return func(c *gin.Context) {

		if autoLogin {
			c.Next()
			return
		}

		session := sessions.Default(c)
		user := session.Get(Userkey)
		if user == nil {
			// Abort the request with the appropriate error code
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// Continue down the chain to handler etc
		c.Next()
	}
}

// Login creates a user session, logging them in given the right username and password
func Login(apiPwd string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := c.PostForm("username")
		password := c.PostForm("password")

		// Validate form input
		if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters can't be empty"})
			return
		}

		// Check for username and password match, usually from a database
		if username != "admin" || password != apiPwd {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			return
		}

		// Save the username in the session
		// set this to the users ID when moving to multi users setup
		session.Set(Userkey, username)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated user"})
	}
}

type accessKey struct {
	AccessKey string `json:"accessKey"`
}

// Cookie Login creates a user session
func CookieLogin(cookiePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		accessKey := accessKey{}

		if err := c.BindJSON(&accessKey); err != nil {
			log.Error().Err(err).Msg("Unable to parse access key from JSON")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse access key from JSON"})
			return
		}

		cookieFile, err := os.ReadFile(cookiePath)
		if err != nil {
			log.Error().Err(err).Msg("Unable to read cookie file")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read cookie file"})
			return
		}

		var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
		cookieFileContents := nonAlphanumericRegex.ReplaceAllString(string(cookieFile), "")

		if subtle.ConstantTimeCompare([]byte(accessKey.AccessKey), []byte(cookieFileContents)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			return
		}

		if err = RefreshCookieFile(cookiePath); err != nil {
			log.Error().Err(err).Msg("Failed to refresh cookie file")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh cookie file"})
			return
		}

		// Save the username in the session
		// set this to the users ID when moving to multi users setup
		session.Set(Userkey, "SSOUser")
		if err := session.Save(); err != nil {
			log.Error().Err(err).Msg("Failed to save session")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Successfully authenticated user"})
	}
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)

	session.Delete(Userkey)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func AutoLoginSetting(autoLogin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, autoLogin)
	}
}
