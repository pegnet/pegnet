package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/pegnet/pegnet/common"

	"github.com/pegnet/pegnet/node/database"

	"github.com/pegnet/pegnet/api"
	log "github.com/sirupsen/logrus"
)

func (n *PegnetNode) NodeAPI(w http.ResponseWriter, r *http.Request) {
	var request api.PostRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		api.Respond(w, api.PostResponse{Err: api.NewJSONDecodingError()})
		return
	}

	n.logger().WithFields(log.Fields{
		"API Method": request.Method,
		"Params":     request.Params}).Info("API Request")

	var result interface{}
	var apiError *api.Error
	switch request.Method {
	case "network-difficulty":
		result, apiError = n.HandleGenericTimeSeries(request.Params, &[]database.DifficultyTimeSeries{})
	case "hash-rate":
		result, apiError = n.HandleGenericTimeSeries(request.Params, &[]database.NetworkHashrateTimeSeries{})
	case "num-records":
		result, apiError = n.HandleGenericTimeSeries(request.Params, &[]database.NumberOPRRecordsTimeSeries{})
	case "asset-price":
		result, apiError = n.HandleGenericTimeSeries(request.Params, &[]database.AssetPricingTimeSeries{})
	case "unique-coinbase":
		result, apiError = n.HandleGenericTimeSeries(request.Params, &[]database.UniqueGradedCoinbasesTimeSeries{})
	case "asset-list":
		result = common.AllAssets
	case "pnt-addresses":
		result = n.PegnetGrader.Balances.AssetHumanReadable("PNT")
	default:
		apiError = api.NewMethodNotFoundError()
	}

	var response api.PostResponse
	if apiError != nil {
		response = api.PostResponse{Err: apiError}
	} else {
		response = api.PostResponse{Res: result}
	}
	api.Respond(w, response)
}

type TimeSeriesParam struct {
	// These params change the response
	AsArray string   `json:"asarray"` // Must choose "time" or "height"
	Values  []string `json:"arrayvalues"`

	// Params that change the query
	DBParams database.FetchTimeSeriesParams `json:"dbparams"`
}

// HandleGenericTimeSeries will handle the time series api call for a given type. It will handle filtering by time
// or block height, converting to a 2d array or array of json objects. It will even handle letting users specify
// the fields in the 2d array format.
//		target		 A type like this &[]database.NetworkHashrateTimeSeries{}
func (n *PegnetNode) HandleGenericTimeSeries(gparams interface{}, target interface{}) (interface{}, *api.Error) {
	params := new(TimeSeriesParam)
	err := api.MapToObject(gparams, params)
	if err != nil {
		return nil, api.NewJSONDecodingError()
	}

	err = database.FetchTimeSeries(n.NodeDatabase.DB, target, &params.DBParams)
	if err != nil {
		e := api.NewInternalError()
		e.Data = err
		return nil, e
	}

	if params.AsArray != "" {
		arr := reflect.Indirect(reflect.ValueOf(target)).Interface()
		return As2DArray(arr, params)
	}

	return target, nil
}

// As2DArray changes the response format from an array of objects to a 2d array for easier
// graph integration.
func As2DArray(rawdata interface{}, params *TimeSeriesParam) ([][]interface{}, *api.Error) {
	rawArray := reflect.ValueOf(rawdata)
	rawArray.Len()

	key := params.AsArray
	switch key {
	case "time", "height", "timems": // Valid
	default: // Invalid
		e := api.NewInvalidParametersError()
		e.Data = fmt.Errorf("'asarray' must be 'time' or 'height'")
		return nil, e
	}

	var result [][]interface{}
	for i := 0; i < rawArray.Len(); i++ {
		item := rawArray.Index(i).Interface().(database.ITimeSeriesData)
		v := []interface{}{dataKey(item, key)}

		for _, value := range params.Values {
			v = append(v, database.FieldValue(item, value))
		}
		result = append(result, v)
	}
	return result, nil
}

func dataKey(data database.ITimeSeriesData, key string) interface{} {
	switch key {
	case "timems":
		return data.Time().Unix() * 1000
	case "time":
		return data.Time().Unix()
	case "height":
		return data.Height()
	}
	// Should never happen
	return -1
}
