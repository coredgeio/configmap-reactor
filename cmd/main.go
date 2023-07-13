package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/coredgeio/configmap-reactor/pkg/config"
	"github.com/coredgeio/configmap-reactor/pkg/reactor"
)

// parseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func parseFlags() string {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "/opt/config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Return the configuration path
	return configPath
}

func main() {
	cfgPath := parseFlags()
	err := config.ParseConfig(cfgPath)
	if err != nil {
		log.Fatalln("unable to parse config", err)
	}
	reactor.CreateConfigMapReactor(config.GetLabel())

	// wait for termination in the main thread, while the main logic
	// keeps running as part of go routines
	sigchan := make(chan os.Signal, 2)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	log.Println("Process Terminated!")
}
