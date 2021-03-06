package app

import (
	"sync"
	"time"

	"github.com/kaedwen/trade/pkg/base"
	"github.com/procyon-projects/chrono"
	"github.com/rs/zerolog/log"
)

var taskScheduler = chrono.NewDefaultTaskScheduler()
var refreshMutex sync.Mutex

func Run() {
	base.Setup()

	quoteChannel := make(chan base.Quote)
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

	// setup alpha vantage api
	SetupCheckQuote(1*time.Minute, 10*time.Second, quoteChannel)

	// block
	signal := make(chan struct{})
	<-signal

}
