// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// KitcoDataSource is the datasource at "https://www.kitco.com/"
type KitcoDataSource struct {
	config *config.Config
}

func NewKitcoDataSource(config *config.Config) (*KitcoDataSource, error) {
	s := new(KitcoDataSource)
	s.config = config
	return s, nil
}

func (d *KitcoDataSource) Name() string {
	return "Kitco"
}

func (d *KitcoDataSource) Url() string {
	return "https://www.kitco.com/"
}

func (d *KitcoDataSource) SupportedPegs() []string {
	return common.CommodityAssets
}

func (d *KitcoDataSource) FetchPegPrices() (peg PegAssets, err error) {
	resp, err := CallKitcoWeb()
	if err != nil {
		return nil, err
	}

	peg = make(map[string]PegItem)

	peg["XAG"], err = d.parseData(resp.Silver)
	if err != nil {
		return nil, err
	}

	peg["XAU"], err = d.parseData(resp.Gold)
	if err != nil {
		return nil, err
	}

	peg["XPD"], err = d.parseData(resp.Palladium)
	if err != nil {
		return nil, err
	}

	peg["XPT"], err = d.parseData(resp.Platinum)
	if err != nil {
		return nil, err
	}

	return
}

func (d *KitcoDataSource) parseData(data KitcoRecord) (PegItem, error) {
	i := PegItem{}
	timestamp, err := time.Parse(d.dateFormat(), data.Date)
	if err != nil {
		return i, err
	}

	v, err := strconv.ParseFloat(data.Bid, 64)
	if err != nil {
		return i, err
	}

	return PegItem{Value: v, When: timestamp, WhenUnix: timestamp.Unix()}, nil
}

func (d *KitcoDataSource) dateFormat() string {
	return "01/02/2006"
}

func (d *KitcoDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}

// ---

type KitcoData struct {
	Silver    KitcoRecord
	Gold      KitcoRecord
	Platinum  KitcoRecord
	Palladium KitcoRecord
	Rhodium   KitcoRecord
}

type KitcoRecord struct {
	Date          string
	Tm            string
	Bid           string
	Ask           string
	Change        string
	PercentChange string
	Low           string
	High          string
}

func CallKitcoWeb() (KitcoData, error) {
	var kData KitcoData
	var emptyData KitcoData

	resp, err := http.Get("https://www.kitco.com/market/")
	if err != nil {
		log.WithError(err).Warning("Failed to get response from Kitco")
		return emptyData, err
	}

	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return emptyData, err
	} else {
		matchStart := "<table class=\"world_spot_price\">"
		matchStop := "</table>"
		strResp := string(body)
		start := strings.Index(strResp, matchStart)
		if start < 0 {
			err = errors.New("No Response")
			log.WithError(err).Warning("Failed to get response from Kitco")
			return emptyData, err
		}
		strResp = strResp[start:]
		stop := strings.Index(strResp, matchStop)
		strResp = strResp[0 : stop+9]
		rows := strings.Split(strResp, "\n")
		for _, r := range rows {
			if strings.Index(r, "wsp-") > 0 {
				ParseKitco(r, &kData)
			}
		}
	}

	return kData, err
}

func ParseKitco(line string, kData *KitcoData) {

	if strings.Index(line, "wsp-AU-date") > 0 {
		kData.Gold.Date = common.PullValue(line, 1)
		//		fmt.Println("kData.Gold.Date:", kData.Gold.Date)
	} else if strings.Index(line, "wsp-AU-time") > 0 {
		kData.Gold.Tm = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-bid") > 0 {
		kData.Gold.Bid = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-ask") > 0 {
		kData.Gold.Ask = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-change") > 0 {
		kData.Gold.Change = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AU-change-percent") > 0 {
		kData.Gold.PercentChange = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AU-low") > 0 {
		kData.Gold.Low = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-high") > 0 {
		kData.Gold.High = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-date") > 0 {
		kData.Silver.Date = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-time") > 0 {
		kData.Silver.Tm = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-bid") > 0 {
		kData.Silver.Bid = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-ask") > 0 {
		kData.Silver.Ask = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-change") > 0 {
		kData.Silver.Change = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AG-change-percent") > 0 {
		kData.Silver.PercentChange = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AG-low") > 0 {
		kData.Silver.Low = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-high") > 0 {
		kData.Silver.High = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-date") > 0 {
		kData.Platinum.Date = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-time") > 0 {
		kData.Platinum.Tm = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-bid") > 0 {
		kData.Platinum.Bid = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-ask") > 0 {
		kData.Platinum.Ask = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-change") > 0 {
		kData.Platinum.Change = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PT-change-percent") > 0 {
		kData.Platinum.PercentChange = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PT-low") > 0 {
		kData.Platinum.Low = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-high") > 0 {
		kData.Platinum.High = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-date") > 0 {
		kData.Palladium.Date = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-time") > 0 {
		kData.Palladium.Tm = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-bid") > 0 {
		kData.Palladium.Bid = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-ask") > 0 {
		kData.Palladium.Ask = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-change") > 0 {
		kData.Palladium.Change = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PD-change-percent") > 0 {
		kData.Palladium.PercentChange = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PD-low") > 0 {
		kData.Palladium.Low = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-high") > 0 {
		kData.Palladium.High = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-date") > 0 {
		kData.Rhodium.Date = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-time") > 0 {
		kData.Rhodium.Tm = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-bid") > 0 {
		kData.Rhodium.Bid = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-ask") > 0 {
		kData.Rhodium.Ask = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-change") > 0 {
		kData.Rhodium.Change = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-RH-change-percent") > 0 {
		kData.Rhodium.PercentChange = common.PullValue(line, 2)
	} else if strings.Index(line, "wsp-RH-low") > 0 {
		kData.Rhodium.Low = common.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-high") > 0 {
		kData.Rhodium.High = common.PullValue(line, 1)
	}
}
