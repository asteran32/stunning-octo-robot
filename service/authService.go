package service

import (
	"app/db"
	"app/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()}) //Err:403
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) //Err:500
		return
	}

	//Create JWT token

	c.JSON(http.StatusOK, user)
}

// SignUp is resgister function
func SignUp(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = db.UserSignUp(user)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
