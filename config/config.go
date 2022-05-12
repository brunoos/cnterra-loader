package config

import (
	"log"
	"os"
	"strconv"
)

var (
	BaseDir = "/opt/cnterra-loader/"
	TmpDir  = "/opt/cnterra-loader/tmp/"

	NodeID     = 0
	DataEx     = "cnterra-data"
	SerialPort = "/dev/ttyUSB0"

	RbAddress  = "localhost"
	RbPort     = "5672"
	RbUser     = "guest"
	RbPassword = "guest"
)

func Initialize() {
	if str, found := os.LookupEnv("NODE_ID"); found {
		n, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			log.Fatal("[ERRO] Invalid 'NODE_ID'")
		}
		NodeID = int(n)
	} else {
		log.Fatal("[ERRO] Variable 'NODE_ID' not set")
	}
	if str, found := os.LookupEnv("SERIAL_PORT"); found {
		SerialPort = str
	}
	if str, found := os.LookupEnv("RABBITMQ_ADDRESS"); found {
		RbAddress = str
	}
	if str, found := os.LookupEnv("RABBITMQ_PORT"); found {
		RbPort = str
	}
	if str, found := os.LookupEnv("RABBITMQ_USER"); found {
		RbUser = str
	}
	if str, found := os.LookupEnv("RABBITMQ_PASSORD"); found {
		RbPassword = str
	}
}