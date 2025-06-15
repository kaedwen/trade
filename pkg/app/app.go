package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kaedwen/trade/pkg/app/client"
	api_error "github.com/kaedwen/trade/pkg/app/error"
	"github.com/kaedwen/trade/pkg/app/session"
	"github.com/kaedwen/trade/pkg/app/utils"
	"github.com/kaedwen/trade/pkg/config"
	"github.com/kaedwen/trade/pkg/model"
)

const (
	ApiAccountPath = "/banking/clients/user/v2/accounts/balances"
)

type Application struct {
	session.Session
	cfg *config.Config
}

func NewApplication() (*Application, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config - %w", err)
	}

	return &Application{session.NewSession(cfg), cfg}, nil
}

func (a *Application) Run(ctx context.Context) error {
	ctx, err := client.NewClient(a.cfg).Connect(ctx)
	if err != nil {
		return err
	}

	if err := a.Init(ctx); err != nil {
		return err
	}

	ctx, err = client.FromContext(ctx).OAuthSecondFlow(ctx)
	if err != nil {
		return err
	}

	t := time.NewTicker(time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			cctx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			if err := a.fetchAccount(cctx); err != nil {
				log.Println("failed to fetch account - %w", err)
			}
		}
	}
}

func (a *Application) fetchAccount(ctx context.Context) error {
	log.Println("running account fetch")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.cfg.ApiAddress.JoinPath(ApiAccountPath).String(), http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Add("x-http-request-info", a.NewRequestInfo())

	resp, err := client.FromContext(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(utils.ReadAllString(resp.Body))
		return api_error.ErrApiBadStatus
	}

	var data model.AccountBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	for _, v := range data.Values {
		log.Println(v.AccountID, v.BalanceEUR, v.AvailableCashAmountEUR)
	}

	return nil
}
