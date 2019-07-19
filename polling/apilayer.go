// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"encoding/json"
	"github.com/zpatrick/go-config"
	"io/ioutil"
	"net/http"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

type APILayerResponse struct {
	Success   bool					`json:"success"`
	Terms     string				`json:"terms"`
	Privacy   string				`json:"privacy"`
	Timestamp int64					`json:"timestamp"`
	Source    string				`json:"source"`
	Quotes    map[string]float64    `json:"quotes"`
	Error     APILayerError			`json:"error"`
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

const apikeyfile = "apikey.dat"

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
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &APILayerResponse)
		return nil
	}

	err = backoff.Retry(operation, PollingExponentialBackOff())
	return APILayerResponse, err
}

func HandleAPILayer(response APILayerResponse, peg *PegAssets) {

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

func APILayerInterface(config *config.Config, peg *PegAssets) {
	log.Debug("Pulling Asset data from APILayer")
	APILayerResponse, err := CallAPILayer(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to access APILayer")
	} else {
		HandleAPILayer(APILayerResponse, peg)
	}
}
