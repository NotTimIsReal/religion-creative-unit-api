package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Start() {
	//get timestamp and store it right now
	readEnv()

	timestamp := time.Now().Unix()
	os.Mkdir("logs", 0777)
	logFile, err := os.OpenFile(fmt.Sprintf("logs/%v.log", timestamp), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)

	}
	defer logFile.Close()
	//combine stdout and log file into one file
	mw := io.MultiWriter(os.Stdout, logFile)
	gin.DefaultWriter = mw
	gin.DefaultErrorWriter = mw
	logger := log.New(mw, "", log.LstdFlags)
	if os.Getenv("MODE") == "release" {
		logger.Println("GIN MODE: release")
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Println("GIN MODE: debug (USING GIN LOGS ONLY)")
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	killed := make(chan bool)
	go func() {
		router.Run(":8080")
		killed <- true
	}()
	<-killed

}
