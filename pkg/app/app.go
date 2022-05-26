package app

import (
	"time"

	"github.com/kaedwen/trade/pkg/base"
	"github.com/procyon-projects/chrono"
	"github.com/rs/zerolog/log"
)

var taskScheduler = chrono.NewDefaultTaskScheduler()

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

	// start influx writer
	go base.RunInflux(dataChannel)

	log.Info().Msg("RUN ...")
	SetupRefreshToken(5*time.Minute, 30*time.Second)
	SetupCheckAccount(1*time.Minute, 20*time.Second, dataChannel)
	SetupCheckDepot(1*time.Minute, 10*time.Second, dataChannel)

	// block
	signal := make(chan struct{})
	<-signal

}
