package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const AuthCookieName = "auth"
const AuthCookieValue = "coredns-webui" // Simple value check for now

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Public paths
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/api/login" {
			c.Next()
			return
		}

		// Check cookie
		cookie, err := c.Cookie(AuthCookieName)
		if err != nil || cookie != AuthCookieValue {
			// If API request, return 401
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			// If Page request, redirect to login
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
