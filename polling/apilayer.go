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
	Success   bool
	Terms     string
	Privacy   string
	Timestamp int64
	Source    string
	Quotes    APILayerRecord
	Error     APILayerError
}

type APILayerError struct {
	Code int64
	Type string
	Info string
}

type APILayerRecord struct {
	USDAED float64
	USDAFN float64
	USDALL float64
	USDAMD float64
	USDANG float64
	USDAOA float64
	USDARS float64
	USDAUD float64
	USDAWG float64
	USDAZN float64
	USDBAM float64
	USDBBD float64
	USDBDT float64
	USDBGN float64
	USDBHD float64
	USDBIF float64
	USDBMD float64
	USDBND float64
	USDBOB float64
	USDBRL float64
	USDBSD float64
	USDBTC float64
	USDBTN float64
	USDBWP float64
	USDBYN float64
	USDBYR float64
	USDBZD float64
	USDCAD float64
	USDCDF float64
	USDCHF float64
	USDCLF float64
	USDCLP float64
	USDCNY float64
	USDCOP float64
	USDCRC float64
	USDCUC float64
	USDCUP float64
	USDCVE float64
	USDCZK float64
	USDDJF float64
	USDDKK float64
	USDDOP float64
	USDDZD float64
	USDEGP float64
	USDERN float64
	USDETB float64
	USDEUR float64
	USDFJD float64
	USDFKP float64
	USDGBP float64
	USDGEL float64
	USDGGP float64
	USDGHS float64
	USDGIP float64
	USDGMD float64
	USDGNF float64
	USDGTQ float64
	USDGYD float64
	USDHKD float64
	USDHNL float64
	USDHRK float64
	USDHTG float64
	USDHUF float64
	USDIDR float64
	USDILS float64
	USDIMP float64
	USDINR float64
	USDIQD float64
	USDIRR float64
	USDISK float64
	USDJEP float64
	USDJMD float64
	USDJOD float64
	USDJPY float64
	USDKES float64
	USDKGS float64
	USDKHR float64
	USDKMF float64
	USDKPW float64
	USDKRW float64
	USDKWD float64
	USDKYD float64
	USDKZT float64
	USDLAK float64
	USDLBP float64
	USDLKR float64
	USDLRD float64
	USDLSL float64
	USDLTL float64
	USDLVL float64
	USDLYD float64
	USDMAD float64
	USDMDL float64
	USDMGA float64
	USDMKD float64
	USDMMK float64
	USDMNT float64
	USDMOP float64
	USDMRO float64
	USDMUR float64
	USDMVR float64
	USDMWK float64
	USDMXN float64
	USDMYR float64
	USDMZN float64
	USDNAD float64
	USDNGN float64
	USDNIO float64
	USDNOK float64
	USDNPR float64
	USDNZD float64
	USDOMR float64
	USDPAB float64
	USDPEN float64
	USDPGK float64
	USDPHP float64
	USDPKR float64
	USDPLN float64
	USDPYG float64
	USDQAR float64
	USDRON float64
	USDRSD float64
	USDRUB float64
	USDRWF float64
	USDSAR float64
	USDSBD float64
	USDSCR float64
	USDSDG float64
	USDSEK float64
	USDSGD float64
	USDSHP float64
	USDSLL float64
	USDSOS float64
	USDSRD float64
	USDSTD float64
	USDSVC float64
	USDSYP float64
	USDSZL float64
	USDTHB float64
	USDTJS float64
	USDTMT float64
	USDTND float64
	USDTOP float64
	USDTRY float64
	USDTTD float64
	USDTWD float64
	USDTZS float64
	USDUAH float64
	USDUGX float64
	USDUSD float64
	USDUYU float64
	USDUZS float64
	USDVEF float64
	USDVND float64
	USDVUV float64
	USDWST float64
	USDXAF float64
	USDXAG float64
	USDXAU float64
	USDXCD float64
	USDXDR float64
	USDXOF float64
	USDXPF float64
	USDYER float64
	USDZAR float64
	USDZMK float64
	USDZMW float64
	USDZWL float64
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

	peg.USD.Value = Round(response.Quotes.USDUSD)
	peg.USD.When = response.Timestamp
	peg.EUR.Value = Round(response.Quotes.USDEUR)
	peg.EUR.When = response.Timestamp
	peg.JPY.Value = Round(response.Quotes.USDJPY)
	peg.JPY.When = response.Timestamp
	peg.GBP.Value = Round(response.Quotes.USDGBP)
	peg.GBP.When = response.Timestamp
	peg.CAD.Value = Round(response.Quotes.USDCAD)
	peg.CAD.When = response.Timestamp
	peg.CHF.Value = Round(response.Quotes.USDCHF)
	peg.CHF.When = response.Timestamp
	peg.INR.Value = Round(response.Quotes.USDINR)
	peg.INR.When = response.Timestamp
	peg.SGD.Value = Round(response.Quotes.USDSGD)
	peg.SGD.When = response.Timestamp
	peg.CNY.Value = Round(response.Quotes.USDCNY)
	peg.CNY.When = response.Timestamp
	peg.HKD.Value = Round(response.Quotes.USDHKD)
	peg.HKD.When = response.Timestamp

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
