package router

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server"
)

func handleRouteUpdate(c *gin.Context) {
	log := c.MustGet("log").(logrus.FieldLogger)

	var wroute models.RouteWrapper

	err := c.BindJSON(&wroute)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	if wroute.Route == nil {
		log.WithError(err).Error(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesMissingNew))
		return
	}

	wroute.Route.AppName = c.Param("app")
	wroute.Route.Name = c.Param("route")

	if err := wroute.Validate(); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, simpleError(err))
		return
	}

	route, err := api.Datastore.StoreRoute(wroute.Route)
	if err != nil {
		log.WithError(err).Debug(models.ErrAppsCreate)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppsCreate))
		return
	}

	c.JSON(http.StatusOK, route)
}
