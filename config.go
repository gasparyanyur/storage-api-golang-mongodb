package main

import (
	"flag"
	"os"
)

const ProtocolVersion = "v1"

type Config struct {
	dbHost     string
	dbName     string
	serverHost string
}

var config Config

func init() {
	flag.StringVar(&config.dbHost, "dbhost", "", "Database host")
	flag.StringVar(&config.dbName, "dbname", "", "Database name")
	flag.StringVar(&config.serverHost, "host", "", "Server host")
	flag.Parse()
	if config.dbHost == "" || config.dbName == "" || config.serverHost == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
}
