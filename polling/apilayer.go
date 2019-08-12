// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type APILayerResponse struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int64              `json:"timestamp"`
	Source    string             `json:"source"`
	Quotes    map[string]float64 `json:"quotes"`
	Error     APILayerError      `json:"error"`
}

type APILayerError struct {
	Code int64
	Type string
	Info string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CallAPILayer(c *config.Config) (response APILayerResponse, err error) {
	var APILayerResponse APILayerResponse

	var apikey string
	{
		apikey, err = c.String("Oracle.APILayerKey")
		check(err)
	}

	operation := func() error {
		resp, err := http.Get("http://www.apilayer.net/api/live?access_key=" + apikey)
		if err != nil {
			log.WithError(err).Warning("Failed to get response from API Layer")
		}

		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else if err = json.Unmarshal(body, &APILayerResponse); err != nil {
			return err
		}
		return err
	}

	err = backoff.Retry(operation, PollingExponentialBackOff())
	// Price is inverted
	if err == nil {
		for k, v := range APILayerResponse.Quotes {
			APILayerResponse.Quotes[k] = 1 / v
		}
	}
	return APILayerResponse, err
}

func HandleAPILayer(response APILayerResponse, peg PegAssets) {

	// Handle Response Errors
	if !response.Success {
		log.WithFields(log.Fields{
			"code": response.Error.Code,
			"type": response.Error.Type,
			"info": response.Error.Info,
		}).Fatal("Failed to access APILayer")
	}

	UpdatePegAssets(response.Quotes, response.Timestamp, peg, "USD")
}

func APILayerInterface(config *config.Config, peg PegAssets) error {
	log.Debug("Pulling Asset data from APILayer")
	APILayerResponse, err := CallAPILayer(config)
	if err != nil {
		return fmt.Errorf("failed to access APILayer : %s", err.Error())
	}
	HandleAPILayer(APILayerResponse, peg)
	return nil
}
