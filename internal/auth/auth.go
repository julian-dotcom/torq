package auth

import (
	"crypto/rand"
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

// AuthRequired is a simple middleware to check the session
func AuthRequired(c *gin.Context) {
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

func Logout(c *gin.Context) {
	session := sessions.Default(c)

	session.Delete(Userkey)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
