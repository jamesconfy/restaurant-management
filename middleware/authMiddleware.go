package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	helper "github.com/jamesconfy/restaurant-management/helpers"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			// c.JSON()
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to do that!"})
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("email", claims.Email)
		c.Set("userId", claims.User_ID)
		c.Set("first_name", claims.First_Name)
		c.Set("last_name", claims.Last_Name)

		c.Next()
	}
}
