package oprecord

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pegnet/OracleRecord/utils"
)

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
	resp, err := http.Get("https://www.kitco.com/market/")
	var kData KitcoData

	if err != nil {
		return kData, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		matchStart := "<table class=\"world_spot_price\">"
		matchStop := "</table>"
		strResp := string(body)
		start := strings.Index(strResp, matchStart)
		if start < 0 {
			return kData, errors.New("No Response")
		}
		strResp = strResp[start:]
		stop := strings.Index(strResp, matchStop)
		strResp = strResp[0 : stop+9]
		rows := strings.Split(strResp, "\n")
		for _, r := range rows {
			if strings.Index(r, "wsp-") > 0 {
				kData = ParseKitco(r, kData)
			}
		}
		//	fmt.Println("RETURNINGKITCO:", strResp)

		return kData, err
	}

}

func ParseKitco(line string, kData KitcoData) KitcoData {

	if strings.Index(line, "wsp-AU-date") > 0 {
		kData.Gold.Date = utils.PullValue(line, 1)
		//		fmt.Println("kData.Gold.Date:", kData.Gold.Date)
	} else if strings.Index(line, "wsp-AU-time") > 0 {
		kData.Gold.Tm = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-bid") > 0 {
		kData.Gold.Bid = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-ask") > 0 {
		kData.Gold.Ask = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-change") > 0 {
		kData.Gold.Change = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AU-change-percent") > 0 {
		kData.Gold.PercentChange = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AU-low") > 0 {
		kData.Gold.Low = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AU-high") > 0 {
		kData.Gold.High = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-date") > 0 {
		kData.Silver.Date = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-time") > 0 {
		kData.Silver.Tm = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-bid") > 0 {
		kData.Silver.Bid = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-ask") > 0 {
		kData.Silver.Ask = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-change") > 0 {
		kData.Silver.Change = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AG-change-percent") > 0 {
		kData.Silver.PercentChange = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-AG-low") > 0 {
		kData.Silver.Low = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-AG-high") > 0 {
		kData.Silver.High = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-date") > 0 {
		kData.Platinum.Date = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-time") > 0 {
		kData.Platinum.Tm = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-bid") > 0 {
		kData.Platinum.Bid = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-ask") > 0 {
		kData.Platinum.Ask = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-change") > 0 {
		kData.Platinum.Change = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PT-change-percent") > 0 {
		kData.Platinum.PercentChange = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PT-low") > 0 {
		kData.Platinum.Low = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PT-high") > 0 {
		kData.Platinum.High = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-date") > 0 {
		kData.Palladium.Date = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-time") > 0 {
		kData.Palladium.Tm = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-bid") > 0 {
		kData.Palladium.Bid = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-ask") > 0 {
		kData.Palladium.Ask = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-change") > 0 {
		kData.Palladium.Change = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PD-change-percent") > 0 {
		kData.Palladium.PercentChange = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-PD-low") > 0 {
		kData.Palladium.Low = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-PD-high") > 0 {
		kData.Palladium.High = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-date") > 0 {
		kData.Rhodium.Date = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-time") > 0 {
		kData.Rhodium.Tm = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-bid") > 0 {
		kData.Rhodium.Bid = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-ask") > 0 {
		kData.Rhodium.Ask = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-change") > 0 {
		kData.Rhodium.Change = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-RH-change-percent") > 0 {
		kData.Rhodium.PercentChange = utils.PullValue(line, 2)
	} else if strings.Index(line, "wsp-RH-low") > 0 {
		kData.Rhodium.Low = utils.PullValue(line, 1)
	} else if strings.Index(line, "wsp-RH-high") > 0 {
		kData.Rhodium.High = utils.PullValue(line, 1)

	}
	return kData
}
