package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaedwen/trade/pkg/base"
	"github.com/procyon-projects/chrono"
	"github.com/rs/zerolog/log"
)

var client = http.DefaultClient
var refreshTokenTask *chrono.ScheduledTask

type SessionResponse struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int    `json:"expires_in"`
	Scope          string `json:"scope"`
	CustomerNumber string `json:"kdnr"`
	BpID           int    `json:"bpid"`
	ContactID      int    `json:"kontaktId"`
}

type SessionStatusResponse struct {
	Identifier       string `json:"identifier"`
	SessionTanActive bool   `json:"sessionTanActive"`
	Activated2FA     bool   `json:"activated2FA"`
}

type SessionValidateResponse struct {
	SessionTanActive bool `json:"sessionTanActive"`
}

type AuthenticationInfo struct {
	ID  string  `json:"id"`
	TYP *string `json:"typ,omitempty"`
}

func OAuthFirstFlow() {

	currentTime := fmt.Sprintf("%d", time.Now().UnixMilli())

	// set our session_id to a new uuid
	base.Config.SessionID = uuid.New().String()
	base.Config.RequestID = currentTime[len(currentTime)-9:]

	data := url.Values{}
	data.Set("client_id", base.Config.ClientID)
	data.Set("client_secret", base.Config.ClientSecret)
	data.Set("grant_type", "password")
	data.Set("username", base.Config.Zugangsnummer)
	data.Set("password", base.Config.Pin)

	req, err := http.NewRequest("POST", base.Config.OAuthUrl+"/oauth/token", strings.NewReader(data.Encode()))
	base.FatalIfError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("OAuthFirstFlow - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data SessionResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		// remember variables
		base.Config.AccessToken = data.AccessToken
		base.Config.RefreshToken = data.RefreshToken

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}

}

func SessionStatus() {

	req, err := http.NewRequest("GET", base.Config.Url+"/session/clients/user/v1/sessions", nil)
	base.FatalIfError(err)

	base.RequestInfo.ClientRequestID = base.ClientRequestID{
		SessionID: base.Config.SessionID,
		RequestID: base.Config.RequestID,
	}

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

	log.Debug().Msgf("SessionStatus - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data []SessionStatusResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		if len(data) > 0 {
			base.Config.SessionUUID = data[0].Identifier
		} else {
			log.Fatal().Msg("Received empty SessionStatusResponse")
		}

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func ValidateSessionTan() {

	data, err := json.Marshal(SessionStatusResponse{
		Identifier:       base.Config.SessionUUID,
		SessionTanActive: true,
		Activated2FA:     true,
	})
	base.FatalIfError(err)

	req, err := http.NewRequest("POST", base.Config.Url+"/session/clients/user/v1/sessions/"+base.Config.SessionUUID+"/validate", bytes.NewBuffer(data))
	base.FatalIfError(err)

	data, err = json.Marshal(base.RequestInfo)
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

	log.Debug().Msgf("ValidateSessionTan - received status code %d", res.StatusCode)

	if res.StatusCode == 201 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data SessionValidateResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		if value := res.Header.Get("x-once-authentication-info"); value != "" {

			var data AuthenticationInfo
			err = json.Unmarshal([]byte(value), &data)
			base.FatalIfError(err)

			// remember id as challenge_id
			base.Config.ChallengeID = data.ID

		} else {
			log.Fatal().Msg("Missing authentication-info Response Header")
		}

	} else {
		log.Fatal().Msgf("Received none 201 Response %d", res.StatusCode)
	}
}

func ActivateSessionTan() {

	data, err := json.Marshal(SessionStatusResponse{
		Identifier:       base.Config.SessionUUID,
		SessionTanActive: true,
		Activated2FA:     true,
	})
	base.FatalIfError(err)

	req, err := http.NewRequest("PATCH", base.Config.Url+"/session/clients/user/v1/sessions/"+base.Config.SessionUUID, bytes.NewBuffer(data))
	base.FatalIfError(err)

	data, err = json.Marshal(base.RequestInfo)
	base.FatalIfError(err)
	req.Header.Set("x-http-request-info", string(data))

	data, err = json.Marshal(AuthenticationInfo{
		ID: base.Config.ChallengeID,
	})
	base.FatalIfError(err)
	req.Header.Set("x-once-authentication-info", string(data))

	req.Header.Set("Authorization", "Bearer "+base.Config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("ActivateSessionTan - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data SessionStatusResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		// remember sessionUUID
		base.Config.SessionUUID = data.Identifier

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func OAuthSecondaryFlow() {

	data := url.Values{}
	data.Set("client_id", base.Config.ClientID)
	data.Set("client_secret", base.Config.ClientSecret)
	data.Set("grant_type", "cd_secondary")
	data.Set("token", base.Config.AccessToken)

	req, err := http.NewRequest("POST", base.Config.OAuthUrl+"/oauth/token", strings.NewReader(data.Encode()))
	base.FatalIfError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("OAuthSecondaryFlow - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data SessionResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		// remember values
		base.Config.AccessToken = data.AccessToken
		base.Config.RefreshToken = data.RefreshToken

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func RefreshToken() {

	data := url.Values{}
	data.Set("client_id", base.Config.ClientID)
	data.Set("client_secret", base.Config.ClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", base.Config.RefreshToken)

	req, err := http.NewRequest("POST", base.Config.OAuthUrl+"/oauth/token", strings.NewReader(data.Encode()))
	base.FatalIfError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	base.FatalIfError(err)

	if res.Body != nil {
		defer res.Body.Close()
	}

	log.Debug().Msgf("RefreshToken - received status code %d", res.StatusCode)

	if res.StatusCode == 200 {

		body, err := ioutil.ReadAll(res.Body)
		base.FatalIfError(err)

		var data SessionResponse
		err = json.Unmarshal(body, &data)
		base.FatalIfError(err)

		log.Trace().Msgf("%+v", data)

		// remember values
		base.Config.AccessToken = data.AccessToken
		base.Config.RefreshToken = data.RefreshToken

	} else {
		log.Fatal().Msgf("Received none 200 Response %d", res.StatusCode)
	}
}

func SetupRefreshToken(rate time.Duration) {
	startTime := time.Now().Add(5 * time.Minute)
	task, err := taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		RefreshToken()
	}, rate, chrono.WithTime(startTime))
	base.FatalIfError(err)

	refreshTokenTask = &task

}

func StopRefreshToken() {
	if refreshTokenTask != nil {
		(*refreshTokenTask).Cancel()
		refreshTokenTask = nil
	}
}
