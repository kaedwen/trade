package model

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
