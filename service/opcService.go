package service

import (
	"app/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ReadCSV(c *gin.Context) {
	csvlist, err := db.GetCSVList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, csvlist)
}
