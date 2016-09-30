package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"

	"github.com/iron-io/functions/api/models"
)

func RunAsyncRunners(mqAdr string) {

	url := fmt.Sprintf("http://%s/tasks", mqAdr)

	logAndWait := func(err error) {
		log.WithError(err)
		time.Sleep(1 * time.Second)
	}

	for {
		resp, err := http.Get(url)
		if err != nil {
			logAndWait(err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logAndWait(err)
			continue
		}

		var task models.Task

		if err := json.Unmarshal(body, &task); err != nil {
			logAndWait(err)
			continue
		}

		if task.ID == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		log.Info("Picked up task:", task.ID)
		var stdout bytes.Buffer                                           // TODO: should limit the size of this, error if gets too big. akin to: https://golang.org/pkg/io/#LimitReader
		stderr := NewFuncLogger(task.RouteName, "", *task.Image, task.ID) // TODO: missing path here, how do i get that?

		if task.Timeout == nil {
			timeout := int32(30)
			task.Timeout = &timeout
		}
		cfg := &Config{
			Image:   *task.Image,
			Timeout: time.Duration(*task.Timeout) * time.Second,
			ID:      task.ID,
			AppName: task.RouteName,
			Stdout:  &stdout,
			Stderr:  stderr,
			Env:     task.EnvVars,
		}

		metricLogger := NewMetricLogger()

		rnr, err := New(metricLogger)
		if err != nil {
			log.WithError(err)
			continue
		}

		ctx := context.Background()
		if _, err = rnr.Run(ctx, cfg); err != nil {
			log.WithError(err)
			continue
		}

		log.Info("Processed task:", task.ID)
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(body))
		if err != nil {
			log.WithError(err)
		}

		c := &http.Client{}
		if _, err := c.Do(req); err != nil {
			log.WithError(err)
			continue
		}

		log.Info("Deleted task:", task.ID)
	}
}
