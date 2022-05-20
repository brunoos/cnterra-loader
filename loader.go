package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/brunoos/cnterra-loader/amqp"
	"github.com/brunoos/cnterra-loader/config"
	"github.com/brunoos/cnterra-loader/serial"
)

//------------------------------------------------------------------------------

type formLoad struct {
	Content string
}

//------------------------------------------------------------------------------

func load(c *gin.Context) {
	serial.Acquire()

	var err error
	var data []byte

	defer func() {
		serial.Release()
		if err == nil {
			go serial.Relay()
		}
	}()

	form := formLoad{}
	err = c.Bind(&form)
	if err != nil {
		log.Println("[ERRO] Error decoding the body:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid body",
		})
		return
	}

	data, err = base64.StdEncoding.DecodeString(form.Content)
	if err != nil {
		log.Println("[ERRO] Error decoding content (base64):", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid body",
		})
		return
	}

	name := config.TmpDir + uuid.NewString() + ".exe"

	err = ioutil.WriteFile(name, data, 0644)
	if err != nil {
		log.Println("[ERRO] Error creating file:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "loading error",
		})
		return
	}

	defer os.Remove(name)
	defer os.Remove(name + ".ihex")
	defer os.Remove("id-" + name + ".ihex")

	bin := config.BaseDir + "loader.sh"
	cmd := exec.Command(bin, config.SerialPort, name)
	err = cmd.Run()
	if err != nil {
		log.Println("[ERRO] Error running loader.sh:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "loading error",
		})
		return
	}

	c.Status(http.StatusOK)
}

func main() {
	config.Initialize()
	amqp.Initialize()

	r := gin.Default()
	r.POST("/load", load)

	log.Println("[INFO] Loader running...")
	r.Run(config.Address + ":" + config.Port)
}
