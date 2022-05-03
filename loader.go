package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.bug.st/serial"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	BaseDir = "/opt/cnterra-loader/"
	TmpDir  = "/opt/cnterra-loader/tmp/"

	ReqDeploy = 100

	ResDeployError = 200
	ResDeployDone  = 201
	ResData        = 202

	SerialFree = iota
	SerialBusy
	SerialWanted
	SerialLoading
)

//------------------------------------------------------------------------------

type RequestBody map[string]interface{}

type ResponseBody map[string]interface{}

type Request struct {
	ID   string      `json:"id"`
	Code int         `json:"code"`
	Body RequestBody `json:"body"`
}

type Response struct {
	ID   string       `json:"id"`
	Code int          `json:"code"`
	Body ResponseBody `json:"body"`
}

//------------------------------------------------------------------------------

var (
	nodeEx     = "node-"
	ctrlEx     = "cnterra-ctrl"
	serialPort = "/dev/ttyUSB0"

	rbUser = "guest"
	rbPass = "guest"
	rbAddr = "cnterra-rabbitmq"
	rbPort = "5672"
)

var channel *amqp.Channel

var mutex sync.Mutex
var cond *sync.Cond = sync.NewCond(&mutex)
var serialStatus int = SerialFree

//------------------------------------------------------------------------------

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//------------------------------------------------------------------------------

func sendResponse(id string, code int, body ResponseBody) error {
	res := Response{
		ID:   id,
		Code: code,
		Body: body,
	}

	buffer, err := json.Marshal(&res)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "text/json",
		Body:        buffer,
	}

	err = channel.Publish(
		ctrlEx, // exchange
		"",     // routing key
		false,  // mandatory
		false,  // immediate
		msg,    // body
	)
	return err
}

//------------------------------------------------------------------------------

func sendData() {
	log.Println("[INFO] Start sending data")

	cond.L.Lock()
	if serialStatus != SerialFree {
		log.Println("[INFO] Serial busy, exiting")
		cond.L.Unlock()
		return
	}
	serialStatus = SerialBusy
	cond.L.Unlock()

	defer func() {
		cond.L.Lock()
		serialStatus = SerialFree
		cond.Signal()
		cond.L.Unlock()
	}()

	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(serialPort, mode)
	if err != nil {
		log.Printf("[ERRO] Error open serial port: %s", err)
		return
	}
	defer port.Close()

	port.SetReadTimeout(50 * time.Millisecond)

	buffer := make([]byte, 1024)
	for {
		n, err := port.Read(buffer)
		if err != nil {
			log.Printf("[ERRO] Error reading data: %s", err)
			return
		}

		if n > 0 {
			err = sendResponse(uuid.NewString(), ResData, ResponseBody{
				"nodeid": 1,
				"data":   base64.StdEncoding.EncodeToString(buffer[:n]),
			})
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

//------------------------------------------------------------------------------

func deploy(req *Request) {
	cond.L.Lock()
	if serialStatus == SerialBusy {
		serialStatus = SerialWanted
		cond.Wait()
	}
	serialStatus = SerialLoading
	cond.L.Unlock()

	var err error
	var data []byte

	defer func() {
		cond.L.Lock()
		serialStatus = SerialFree
		cond.L.Unlock()

		if err == nil {
			go sendData()
		}
	}()

	content := req.Body["content"]
	data, err = base64.StdEncoding.DecodeString(content.(string))

	if err != nil {
		log.Printf("[ERRO] Error decoding content (base64): %s", err)
		sendResponse(req.ID, ResDeployError, ResponseBody{"message": "Error decoding content (base64)"})
		return
	}

	name := TmpDir + uuid.NewString() + ".exe"

	err = ioutil.WriteFile(name, data, 0644)
	if err != nil {
		log.Printf("[ERRO] Error creating file: %s", err)
		sendResponse(req.ID, ResDeployError, ResponseBody{"message": "Error deploying content"})
		return
	}

	defer os.Remove(name)
	defer os.Remove(name + ".ihex")
	defer os.Remove("id-" + name + ".ihex")

	bin := BaseDir + "loader.sh"
	cmd := exec.Command(bin, serialPort, name)
	err = cmd.Run()
	if err != nil {
		log.Printf("[ERRO] Error running loader.sh: %s", err)
		sendResponse(req.ID, ResDeployError, ResponseBody{"message": "Error deploying content"})
		return
	}

	err = sendResponse(req.ID, ResDeployDone, ResponseBody{"message": "Deploy done"})
	if err != nil {
		log.Printf("[ERRO] Error sending response: %s", err)
		return
	}
}

func dispatch(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		req := Request{}

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			log.Printf("[ERRO] Error decoding request: %s", err)
			continue
		}

		if req.Code != ReqDeploy {
			continue
		}

		deploy(&req)
	}
}

//------------------------------------------------------------------------------

func getEnv() {
	if str, found := os.LookupEnv("NODE_ID"); !found {
		log.Fatal("[ERRO] Variable 'NODE_ID' not found")
	} else {
		nodeEx = "node-" + str
	}
	if str, found := os.LookupEnv("SERIAL_PORT"); found {
		serialPort = str
	}
	if str, found := os.LookupEnv("RABBITMQ_ADDRESS"); found {
		rbAddr = str
	}
	if str, found := os.LookupEnv("RABBITMQ_PORT"); found {
		rbPort = str
	}
	if str, found := os.LookupEnv("RABBITMQ_USER"); found {
		rbUser = str
	}
	if str, found := os.LookupEnv("RABBITMQ_PASS"); found {
		rbPort = str
	}
}

func main() {
	getEnv()

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", rbUser, rbPass, rbAddr, rbPort)
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		nodeEx,   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	err = channel.ExchangeDeclare(
		ctrlEx,   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	queue, err := channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.QueueBind(
		queue.Name, // queue name
		"",         // routing key
		nodeEx,     // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	failOnError(err, "Failed to register a consumer")

	var wg sync.WaitGroup
	wg.Add(1)
	go dispatch(msgs)

	log.Println("[INFO] Loader running...")

	wg.Wait()
}
