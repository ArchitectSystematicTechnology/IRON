/*

For keeping a minimum running, perhaps when doing a routing table update, if destination hosts are all
 expired or about to expire we start more.

*/

package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server"
)

func main() {
	config := &models.Config{}
	config.DatabaseURL = os.Getenv("DB")
	config.API = os.Getenv("API")
	config.PapertrailToken = os.Getenv("PT_TOKEN")

	err := config.Validate()
	if err != nil {
		log.WithError(err).Fatalln("Invalid config.")
	}
	log.Printf("config: %+v", config)

	api := server.New(config)
	api.Start()
}
