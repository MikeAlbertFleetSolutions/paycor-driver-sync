package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/MikeAlbertFleetSolutions/paycor-driver-sync/config"
	"github.com/MikeAlbertFleetSolutions/paycor-driver-sync/paycor"
)

var (
	buildnum string
)

func main() {
	// show file & location, date & time
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// command line app
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s build %s\n", os.Args[0], buildnum)
		flag.PrintDefaults()
	}

	// process command line
	var configFile string
	flag.StringVar(&configFile, "config", "", "Configuration file")
	flag.Parse()

	if len(configFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// read config
	err := config.FromFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// create paycor client
	pc, err := paycor.NewClient(config.Paycor.PublicKey, config.Paycor.PrivateKey, config.Paycor.Host)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// get employees to sync over
	drivers, err := pc.GetDriverHomeAddresses(config.Paycor.HomeAddressesReport)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	for _, driver := range drivers {
		log.Printf("Driver %+v", driver)
	}
}
