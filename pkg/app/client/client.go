package client

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	api_error "github.com/kaedwen/trade/pkg/app/error"
	"github.com/kaedwen/trade/pkg/app/utils"
	"github.com/kaedwen/trade/pkg/config"
	"golang.org/x/oauth2"
)

type Client interface {
	Connect(context.Context) (context.Context, error)
	OAuthSecondFlow(ctx context.Context) (context.Context, error)
	Do(*http.Request, ...ClientOption) (*http.Response, error)
}

type ClientOption func(*clientOptions)
type ClientOptions []ClientOption

type clientOptions struct {
	retryStatusCodes []int
	retryDelay       time.Duration
	retryMax         int
}

type oauthSecondaryFlowResponse struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int64  `json:"expires_in"`
	Scope          string `json:"scope"`
	CustomerNumber string `json:"kdnr"`
	BpID           int    `json:"bpid"`
	ContactID      int    `json:"kontaktId"`
}

func WithRetryCode(code int) ClientOption {
	return func(co *clientOptions) {
		co.retryStatusCodes = append(co.retryStatusCodes, code)
	}
}

func WithRetryDelay(d time.Duration) ClientOption {
	return func(co *clientOptions) {
		co.retryDelay = d
	}
}

func WithRetryMax(max int) ClientOption {
	return func(co *clientOptions) {
		co.retryMax = max
	}
}

type client struct {
	*http.Client
	cfg *config.Config
	oac *oauth2.Config
	tks oauth2.TokenSource
}

func NewClient(cfg *config.Config) Client {
	oac := &oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: cfg.TokenAddress.JoinPath("oauth/token").String(),
		},
	}

	return &client{cfg: cfg, oac: oac}
}

func (c *client) Connect(ctx context.Context) (context.Context, error) {
	log.Println("oauth flow")

	tk, err := c.oac.PasswordCredentialsToken(ctx, c.cfg.AccountId, c.cfg.Pin)
	if err != nil {
		return nil, err
	}

	c.tks = c.oac.TokenSource(ctx, tk)
	c.Client = &http.Client{Transport: &jsonTransport{oauth2.NewClient(ctx, c.tks).Transport}}
	return contextWithClient(ctx, c), nil
}

func (c *client) OAuthSecondFlow(ctx context.Context) (context.Context, error) {
	log.Println("oauth cd_secondary flow")

	tk, err := c.tks.Token()
	if err != nil {
		return nil, err
	}

	data := url.Values{
		"client_id":     {c.oac.ClientID},
		"client_secret": {c.oac.ClientSecret},
		"grant_type":    {"cd_secondary"},
		"token":         {tk.AccessToken},
	}

	req, err := http.NewRequest(http.MethodPost, c.oac.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(utils.ReadAllString(resp.Body))
		return nil, api_error.ErrApiBadStatus
	}

	var oauthSecondaryFlowResponse oauthSecondaryFlowResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthSecondaryFlowResponse); err != nil {
		return nil, err
	}

	stk := &oauth2.Token{
		AccessToken:  oauthSecondaryFlowResponse.AccessToken,
		RefreshToken: oauthSecondaryFlowResponse.RefreshToken,
		ExpiresIn:    oauthSecondaryFlowResponse.ExpiresIn,
	}

	c.tks = c.oac.TokenSource(ctx, stk)
	c.Client = &http.Client{Transport: &jsonTransport{oauth2.NewClient(ctx, c.tks).Transport}}
	return contextWithClient(ctx, c), nil
}

func (c *client) Do(req *http.Request, opt ...ClientOption) (resp *http.Response, err error) {
	opts := clientOptions{
		retryDelay: time.Second,
		retryMax:   10,
	}

	for _, o := range opt {
		o(&opts)
	}

	for {
		resp, err = c.Client.Do(req.Clone(req.Context()))
		if err != nil {
			return
		}

		if slices.Contains(opts.retryStatusCodes, resp.StatusCode) {
			time.Sleep(opts.retryDelay)

			opts.retryMax--
			if opts.retryMax > 0 {
				continue
			} else {
				return resp, errors.New("retry failed")
			}
		}

		return
	}
}

type jsonTransport struct {
	http.RoundTripper
}

func (jt *jsonTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Add("Content-Type", "application/json")
	req2.Header.Add("Accept", "application/json")

	return jt.RoundTripper.RoundTrip(req2)
}

type contextKey string

const clientContextKey contextKey = "client-context"

func contextWithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, clientContextKey, c)
}

func FromContext(ctx context.Context) Client {
	return ctx.Value(clientContextKey).(Client)
}
