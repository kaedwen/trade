package app

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kaedwen/trade/pkg/base"
	"github.com/procyon-projects/chrono"
	"github.com/rs/zerolog/log"
)

var checkDepotTask *chrono.ScheduledTask

type UnitValue struct {
	Value float64 `json:"value,string"`
	Unit  string  `json:"unit"`
}

type DepotDetails struct {
	DepotID                    string   `json:"depotId"`
	DepotDisplayID             string   `json:"depotDisplayId"`
	ClientID                   string   `json:"clientId"`
	DepotType                  string   `json:"depotType"`
	DefaultSettlementAccountId string   `json:"defaultSettlementAccountId"`
	SettlementAccountIds       []string `json:"settlementAccountIds"`
	HolderName                 string   `json:"holderName"`
}

type DepotReponse struct {
	Paging struct {
		Index   int `json:"index"`
		Matches int `json:"matches"`
	} `json:"paging"`
	Values []DepotDetails `json:"values"`
}

type DepotDetailsResponse struct {
	Paging struct {
		Index   int `json:"index"`
		Matches int `json:"matches"`
	} `json:"paging"`
	Aggregated struct {
		Depot                 DepotDetails `json:"depot"`
		PrevDayValue          UnitValue    `json:"prevDayValue"`
		CurrentValue          UnitValue    `json:"currentValue"`
		PurchaseValue         UnitValue    `json:"purchaseValue"`
		ProfitLossPurchaseAbs UnitValue    `json:"profitLossPurchaseAbs"`
		ProfitLossPrevDayAbs  UnitValue    `json:"profitLossPrevDayAbs"`
		ProfitLossPurchaseRel float64      `json:"profitLossPurchaseRel,string"`
		ProfitLossPrevDayRel  float64      `json:"profitLossPrevDayRel,string"`
	} `json:"aggregated"`
	Values []struct {
		DepotID           string    `json:"depotId"`
		PositionID        string    `json:"positionId"`
		WKN               string    `json:"wkn"`
		CustodyType       string    `json:"custodyType"`
		Quantity          UnitValue `json:"quantity"`
		AvailableQuantity UnitValue `json:"availableQuantity"`
		CurrentPrice      struct {
			Price         UnitValue `json:"price"`
			PriceDateTime string    `json:"priceDateTime"`
		}
		PurchasePrice UnitValue `json:"purchasePrice"`
		PrevDayPrice  struct {
			Price         UnitValue `json:"price"`
			PriceDateTime string    `json:"priceDateTime"`
		}
		CurrentValue             UnitValue `json:"currentValue"`
		PurchaseValue            UnitValue `json:"purchaseValue"`
		ProfitLossPurchaseAbs    UnitValue `json:"profitLossPurchaseAbs"`
		ProfitLossPrevDayAbs     UnitValue `json:"profitLossPrevDayAbs"`
		AvailableQuantityToHedge UnitValue `json:"availableQuantityToHedge"`
		ProfitLossPurchaseRel    float64   `json:"profitLossPurchaseRel,string"`
		ProfitLossPrevDayRel     float64   `json:"profitLossPrevDayRel,string"`
		Version                  string    `json:"version"`
		Hedgeability             string    `json:"hedgeability"`
		CurrentPriceDeterminable bool      `json:"currentPriceDeterminable"`
	} `json:"values"`
}

func queryDepot() []string {

	req, err := http.NewRequest("GET", base.Config.Comdirect.Url+"/brokerage/clients/user/v3/depots", nil)
	base.FatalIfError(err)

	data, err := json.Marshal(base.RequestInfo)
	base.FatalIfError(err)

	req.Header.Set("Authorization", "Bearer "+base.Config.AccessToken)
	req.Header.Set("x-http-request-info", string(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("checkDepot - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data DepotReponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		var list []string

		for _, item := range data.Values {
			list = append(list, item.DepotID)
		}

		return list

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}

	return nil
}

func checkDepotDetails(depotID string, dataChannel chan base.Data) {

	req, err := http.NewRequest("GET", base.Config.Comdirect.Url+"/brokerage/v3/depots/"+depotID+"/positions", nil)
	base.FatalIfError(err)

	data, err := json.Marshal(base.RequestInfo)
	base.FatalIfError(err)

	req.Header.Set("Authorization", "Bearer "+base.Config.AccessToken)
	req.Header.Set("x-http-request-info", string(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("checkDepotDetails - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data DepotDetailsResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		for _, item := range data.Values {
			dataChannel <- base.Data{
				Name:   "Depot",
				Tags:   map[string]string{"WKN": item.WKN},
				Fields: map[string]interface{}{"Value": item.CurrentValue.Value},
			}
		}

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func SetupCheckDepot(rate time.Duration, delay time.Duration, dataChannel chan base.Data) {
	depotIDs := queryDepot()
	task, err := taskScheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		refreshMutex.Lock()
		defer refreshMutex.Unlock()
		checkDepotDetails(depotIDs[0], dataChannel)
	}, rate, chrono.WithTime(time.Now().Add(delay)))
	base.FatalIfError(err)

	checkDepotTask = &task

}

func StopCheckDepot() {
	if checkDepotTask != nil {
		(*checkDepotTask).Cancel()
		checkDepotTask = nil
	}
}
