package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type message struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
type posts struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Type    string `json:"type"`
	Content string `json:"content"`
}
type Main struct {
	Killed chan bool
}

func (m *Main) Start() {
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
	logger.Println("Inititalizing MongoDB connection")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client()
	clientOptions = clientOptions.ApplyURI(os.Getenv("MONGO_URI")).SetServerAPIOptions(serverAPIOptions).SetMaxPoolSize(50).SetCompressors([]string{"zlib"}).SetRetryWrites(true)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	logger.Println("Connected to MongoDB")
	router := gin.New()
	router.Use(gin.Logger())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, message{
			Message: "pong",
			Status:  200,
		})
	})
	router.GET("/posts", func(c *gin.Context) {
		data, err := client.Database("religion-creative").Collection("posts").Find(ctx, bson.D{})
		var posts []posts

		if err != nil {
			//make posts an empty array
			logger.Println("[ERROR] Failed to get posts from MongoDB: ", err)
			c.JSON(200, []string{})
			return
		} else {

			err = data.All(ctx, &posts)
			if err != nil {
				logger.Println("[ERROR] Failed to decode posts from MongoDB: ", err)
				logger.Println("[ERROR] Posts: ", posts)
				c.Status(500)
				c.JSON(500, message{
					Message: "Internal Server Error",
					Status:  500,
				})
				return
			}
		}

		c.JSON(200, posts)
	})
	router.POST("/posts", func(c *gin.Context) {
		//get post data
		var post posts
		err := c.BindJSON(&post)
		if err != nil {
			logger.Println("[ERROR] Failed to get post data from request: ", err)
			c.Status(400)
			c.JSON(400, message{
				Message: "Bad Request",
				Status:  400,
			})
			return
		}
		//insert post data into database
		result, err := client.Database("religion-creative").Collection("posts").InsertOne(ctx, post)
		if err != nil {
			logger.Println("[ERROR] Failed to insert post data into MongoDB: ", err)
			c.Status(500)
			c.JSON(500, message{
				Message: "Internal Server Error",
				Status:  500,
			})
			return
		}
		//return post data
		c.JSON(201, result)
	})
	router.DELETE("/posts", func(c *gin.Context) {
		//get query name
		title := c.Query("title")
		author := c.Query("author")
		if title == "" || author == "" {
			logger.Println("[ERROR] Failed to get title from query")
			c.Status(400)
			c.JSON(400, message{
				Message: "Bad Request",
				Status:  400,
			})
			return
		}

		//delete post
		result, err := client.Database("religion-creative").Collection("posts").DeleteOne(ctx, bson.M{"title": title, "author": author})
		if err != nil {
			logger.Println("[ERROR] Failed to delete post from MongoDB: ", err)
			c.Status(500)
			c.JSON(500, message{
				Message: "Internal Server Error",
				Status:  500,
			})
			return
		}
		//return post data
		c.JSON(200, result)

	})
	router.POST("/ghpayload", func(c *gin.Context) {
		c.Status(200)
		c.JSON(200, message{
			Message: "OK",
			Status:  200,
		})
		//wait once second to make sure the response is sent
		go func() {
			time.Sleep(time.Second)
			m.End()
		}()
		return

	})
	go func() {
		router.Run(":8080")
		m.Killed <- true

	}()
	<-m.Killed
	logger.Println("Shutting down server")

}
func (m *Main) End() {
	m.Killed <- true

}
