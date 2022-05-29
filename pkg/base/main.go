package base

func Setup() {

	// first setup config
	setupConfig()

	// setup logger
	setupLogger()

	// setup finnhub
	setupAlpha()

	// setup Influx connection
	setupInflux()
}
