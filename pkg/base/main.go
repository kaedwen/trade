package base

func Setup() {

	// first setup config
	setupConfig()

	// setup logger
	setupLogger()

	// setup Influx connection
	setupInflux()
}
