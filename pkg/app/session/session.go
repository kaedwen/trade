package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kaedwen/trade/pkg/app/client"
	api_error "github.com/kaedwen/trade/pkg/app/error"
	"github.com/kaedwen/trade/pkg/app/utils"
	"github.com/kaedwen/trade/pkg/config"
)

const (
	ApiSessionUserPath     = "/session/clients/user/v1/sessions"
	ApiSessionValidatePath = "/session/clients/user/v1/sessions/%s/validate"
	ApiSessionActivatePath = "/session/clients/user/v1/sessions/%s"
)

type Session interface {
	Init(context.Context) error
	NewRequestInfo() string
}

type session struct {
	cfg         *config.Config
	sessionId   string
	challengeId string
}

type sessionData struct {
	Identifier       string `json:"identifier"`
	SessionTanActive bool   `json:"sessionTanActive"`
	Activated2FA     bool   `json:"activated2FA"`
}

type sessionAuthenticationInfo struct {
	Id             string   `json:"id,omitempty"`
	Type           *string  `json:"typ,omitempty"`
	Challenge      string   `json:"challenge,omitempty"`
	AvailableTypes []string `json:"availableTypes,omitempty"`
}

type requestInfo struct {
	ClientRequestId clientRequestId `json:"clientRequestId"`
}

type clientRequestId struct {
	SessionId string `json:"sessionId"`
	RequestId string `json:"requestId"`
}

func NewSession(cfg *config.Config) Session {
	return &session{
		cfg: cfg,
	}
}

func (s *session) Init(ctx context.Context) error {
	if err := s.aquireSession(ctx); err != nil {
		return err
	}

	if err := s.validateSessionTan(ctx); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)

	log.Println("wait 30s for TAN challenge (or SIGHUP) ...")
	select {
	case <-signals:
	case <-time.After(30 * time.Second):
	}

	close(signals)

	if err := s.activateSession(ctx); err != nil {
		return err
	}

	return nil
}

func (s *session) Id() string {
	return s.sessionId
}

func (s *session) aquireSession(ctx context.Context) error {
	log.Println("aquire session")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.ApiAddress.JoinPath(ApiSessionUserPath).String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("x-http-request-info", newRequestInfo(utils.RandString(100)))

	resp, err := client.FromContext(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(utils.ReadAllString(resp.Body))
		return api_error.ErrApiBadStatus
	}

	var sessionDataList []sessionData
	if err := json.NewDecoder(resp.Body).Decode(&sessionDataList); err != nil {
		return err
	}

	if len(sessionDataList) == 0 {
		return errors.New("no session objects received")
	}

	s.sessionId = sessionDataList[0].Identifier

	return nil
}

func (s *session) validateSessionTan(ctx context.Context) error {
	log.Println("validate session")

	data, _ := json.Marshal(sessionData{
		Identifier:       s.sessionId,
		SessionTanActive: true,
		Activated2FA:     true,
	})

	req, err := http.NewRequest(http.MethodPost, s.cfg.ApiAddress.JoinPath(fmt.Sprintf(ApiSessionValidatePath, s.sessionId)).String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("x-http-request-info", s.NewRequestInfo())

	resp, err := client.FromContext(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Println(utils.ReadAllString(resp.Body))
		return api_error.ErrApiBadStatus
	}

	value := resp.Header.Get("x-once-authentication-info")
	if len(value) == 0 {
		return errors.New("no authentication-info received")
	}

	var authenticationInfo sessionAuthenticationInfo
	if err := json.Unmarshal([]byte(value), &authenticationInfo); err != nil {
		return errors.New("failed to parse authentication-info")
	}

	s.challengeId = authenticationInfo.Id

	return nil
}

func (s *session) activateSession(ctx context.Context) error {
	log.Println("activate session")

	data, _ := json.Marshal(sessionData{
		Identifier:       s.sessionId,
		SessionTanActive: true,
		Activated2FA:     true,
	})

	req, err := http.NewRequest(http.MethodPatch, s.cfg.ApiAddress.JoinPath(fmt.Sprintf(ApiSessionActivatePath, s.sessionId)).String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("x-http-request-info", s.NewRequestInfo())
	req.Header.Add("x-once-authentication-info", newAuthenticationInfo(s.challengeId))

	resp, err := client.FromContext(ctx).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(utils.ReadAllString(resp.Body))
		return api_error.ErrApiBadStatus
	}

	return nil
}

func (s *session) NewRequestInfo() string {
	return newRequestInfo(s.sessionId)
}

func newAuthenticationInfo(challengeId string) string {
	ri := sessionAuthenticationInfo{
		Id: challengeId,
	}

	rid, _ := json.Marshal(ri)

	return string(rid)
}

func newRequestInfo(sessionId string) string {
	ri := requestInfo{
		ClientRequestId: clientRequestId{
			SessionId: sessionId,
			RequestId: utils.RandString(25),
		},
	}

	rid, _ := json.Marshal(ri)

	return string(rid)
}
