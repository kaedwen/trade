package base

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
)

var writeAPI api.WriteAPIBlocking

func setupInflux() {

	client := influxdb2.NewClient(Config.Influx.Url, Config.Influx.Token)
	writeAPI = client.WriteAPIBlocking(Config.Influx.Org, Config.Influx.Bucket)

}

func RunInflux(dataChannel chan Data) {
	for data := range dataChannel {
		err := writeAPI.WritePoint(context.Background(), influxdb2.NewPoint(data.Name,
			data.Tags,
			data.Fields,
			time.Now()))
		FatalIfError(err)
	}
	log.Error().Msg("For some reason RunInflux will die!")
}
