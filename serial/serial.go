package serial

import (
	"encoding/base64"
	"log"
	"sync"
	"time"

	"github.com/brunoos/cnterra-loader/amqp"
	"github.com/brunoos/cnterra-loader/config"
	"go.bug.st/serial"
)

const (
	SerialFree = iota
	SerialData
	SerialWanted
	SerialLoad
)

//------------------------------------------------------------------------------

var mutex sync.Mutex
var cond *sync.Cond = sync.NewCond(&mutex)
var serialStatus int = SerialFree

//------------------------------------------------------------------------------

// Acquire serial port for loading
func Acquire() {
	cond.L.Lock()
	if serialStatus == SerialData {
		serialStatus = SerialWanted
		cond.Wait()
	}
	serialStatus = SerialLoad
	cond.L.Unlock()
}

// Release the serial port
func Release() {
	cond.L.Lock()
	serialStatus = SerialFree
	cond.Signal()
	cond.L.Unlock()
}

//------------------------------------------------------------------------------

func Relay() {
	log.Println("[INFO] Start to relay serial data")

	cond.L.Lock()
	if serialStatus != SerialFree {
		log.Println("[INFO] Serial busy, exiting")
		cond.L.Unlock()
		return
	}
	serialStatus = SerialData
	cond.L.Unlock()

	defer func() {
		Release()
	}()

	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(config.SerialPort, mode)
	if err != nil {
		log.Printf("[ERRO] Error open serial port: %s", err)
		return
	}
	defer port.Close()

	port.SetReadTimeout(100 * time.Millisecond)

	buffer := make([]byte, 1024)
	for {
		n, err := port.Read(buffer)
		if err != nil {
			log.Printf("[ERRO] Error reading data: %s", err)
			return
		}

		if n > 0 {
			data := base64.StdEncoding.EncodeToString(buffer[:n])
			err = amqp.SendData(data)
			if err != nil {
				log.Printf("[ERRO] Error sending data: %s", err)
				return
			}
		}

		if serialStatus == SerialWanted {
			log.Println("[INFO] Serial wanted, disconnecting")
			return
		}
	}
}
