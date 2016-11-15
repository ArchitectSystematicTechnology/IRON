package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Version of IronFunctions
var Version = "0.0.76"

func handleVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": Version})
}
