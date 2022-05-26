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

var checkAccountTask *chrono.ScheduledTask

type AccountBalanceResponse struct {
	Paging struct {
		Index   int `json:"index"`
		Matches int `json:"matches"`
	} `json:"paging"`
	Values []struct {
		AccountID string `json:"accountId"`
		Account   struct {
			AccountId        string `json:"accountId"`
			AccountDisplayId string `json:"accountDisplayId"`
			Currency         string `json:"currency"`
			ClientID         string `json:"clientId"`
			IBAN             string `json:"iban"`
			AccountType      struct {
				Key  string `json:"key"`
				Text string `json:"text"`
			} `json:"accountType"`
			CreditLimit struct {
				Value float64 `json:"value,string"`
				Unit  string  `json:"unit"`
			} `json:"creditLimit"`
		} `json:"account"`
		Balance struct {
			Value float64 `json:"value,string"`
			Unit  string  `json:"unit"`
		} `json:"balance"`
		BalanceEUR struct {
			Value float64 `json:"value,string"`
			Unit  string  `json:"unit"`
		} `json:"balanceEUR"`
		AvailableCashAmount struct {
			Value float64 `json:"value,string"`
			Unit  string  `json:"unit"`
		} `json:"availableCashAmount"`
		AvailableCashAmountEUR struct {
			Value float64 `json:"value,string"`
			Unit  string  `json:"unit"`
		} `json:"availableCashAmountEUR"`
	} `json:"values"`
}

func checkAccount(dataChannel chan base.Data) {

	req, err := http.NewRequest("GET", base.Config.Url+"/banking/clients/user/v2/accounts/balances", nil)
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

	log.Debug().Msgf("checkAccount - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data AccountBalanceResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		for _, item := range data.Values {
			dataChannel <- base.Data{
				Name:   "Account",
				Tags:   map[string]string{"IBAN": item.Account.IBAN},
				Fields: map[string]interface{}{"Value": item.BalanceEUR.Value},
			}
		}

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func SetupCheckAccount(rate time.Duration, delay time.Duration, dataChannel chan base.Data) {
	task, err := taskScheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		checkAccount(dataChannel)
	}, rate, chrono.WithTime(time.Now().Add(delay)))
	base.FatalIfError(err)

	checkAccountTask = &task

}

func StopCheckAccount() {
	if checkAccountTask != nil {
		(*checkAccountTask).Cancel()
		checkAccountTask = nil
	}
}
