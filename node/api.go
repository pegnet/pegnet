package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

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
		result, apiError = n.HandleDifficultyTimeSeries(request.Params)
	case "hash-rate":
		result, apiError = n.HandleNetworkHashRateTimeSeries(request.Params)
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

//
//func (a *APIServer) getPerformance(params interface{}) (*PerformanceResult, *Error) {
//	performanceParams := new(PerformanceParameters)
//	err := MapToObject(params, performanceParams)
//	if err != nil {
//		return nil, NewJSONDecodingError()
//	}

type TimeSeriesParam struct {
	// These params change the response
	AsArray string   `json:"asarray"` // Must choose "time" or "height"
	Values  []string `json:"arrayvalues"`

	// Params that change the query
	DBParams database.FetchTimeSeriesParams `json:"dbparams"`
}

func (n *PegnetNode) HandleDifficultyTimeSeries(gparams interface{}) (interface{}, *api.Error) {
	params := new(TimeSeriesParam)
	err := api.MapToObject(gparams, params)
	if err != nil {
		return nil, api.NewJSONDecodingError()
	}

	var diffs interface{} = &[]database.DifficultyTimeSeries{}
	err = database.FetchTimeSeries(n.NodeDatabase.DB, diffs, &params.DBParams)
	if err != nil {
		e := api.NewInternalError()
		e.Data = err
		return nil, e
	}

	data := diffs.(*[]database.DifficultyTimeSeries)

	if params.AsArray != "" {
		return AsArray(*data, params)
	}

	return *data, nil
}

func (n *PegnetNode) HandleNetworkHashRateTimeSeries(gparams interface{}) (interface{}, *api.Error) {
	params := new(TimeSeriesParam)
	err := api.MapToObject(gparams, params)
	if err != nil {
		return nil, api.NewJSONDecodingError()
	}

	var diffs interface{} = &[]database.NetworkHashrateTimeSeries{}
	err = database.FetchTimeSeries(n.NodeDatabase.DB, diffs, &params.DBParams)
	if err != nil {
		e := api.NewInternalError()
		e.Data = err
		return nil, e
	}

	data := diffs.(*[]database.NetworkHashrateTimeSeries)

	if params.AsArray != "" {
		return AsArray(*data, params)
	}

	return *data, nil
}

func AsArray(rawdata interface{}, params *TimeSeriesParam) ([][]interface{}, *api.Error) {
	rawArray := reflect.ValueOf(rawdata)
	rawArray.Len()

	key := params.AsArray
	switch key {
	case "time", "height": // Valid
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
	case "time":
		return data.Time().Unix()
	case "height":
		return data.Height()
	}
	// Should never happen
	return -1
}
