package service

import (
	"app/db"
	"app/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SignIn is login function
func SignIn(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err = db.UserSignIn(user.Email, user.Password)
	if err != nil {
		if err == db.ErrINVALIDEMAIL {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Generate token
	tokens, err := GenerateJWT(user.Email)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tokens)
}

// SignUp is resgister function
func SignUp(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Database user info insert
	err = db.UserSignUp(user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// SignOut is logout function
func SignOut(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization Token is required"})
		return
	}
	extracted := strings.Split(tokenStr, "Bearer ")
	if len(extracted) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Forat of Authorization Token"})
		return
	}
	// Delete Access token
	if err := DeleteToken(extracted[1]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// User page
func User(c *gin.Context) {
	tokenStr := c.GetHeader("Authorization")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization Token is required"})
		return
	}
	extracted := strings.Split(tokenStr, "Bearer ")
	if len(extracted) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Forat of Authorization Token"})
		return
	}
	// check token vaild
	if err := VerifyAccessToken(extracted[1]); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Refresh
func Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	//
	refreshToken := mapToken["refresh_token"]
	if err := VerifyRefreshToken(refreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "refresh token is expired"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success refresh"})
}
