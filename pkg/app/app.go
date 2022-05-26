package app

import (
	"time"

	"github.com/kaedwen/trade/pkg/base"
	"github.com/rs/zerolog/log"
)

func Run() {
	base.Setup()

	dataChannel := make(chan base.Data)

	OAuthFirstFlow()
	SessionStatus()
	ValidateSessionTan()

	log.Info().Msg("Wait 30s for TAN Challenge ...")
	time.Sleep(30 * time.Second)

	ActivateSessionTan()
	OAuthSecondaryFlow()

	signal := make(chan struct{})

	// start influx writer
	go base.RunInflux(dataChannel)

	log.Info().Msg("RUN ...")
	SetupRefreshToken(5 * time.Minute)
	SetupCheckAccount(1*time.Minute, dataChannel)

	<-signal

}
