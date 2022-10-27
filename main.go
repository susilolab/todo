package main

import (
	"log"
	"os"
	"runtime"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"todo/models"
	"todo/services"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	dbURL := os.Getenv("DB_PATH")
	models.SetupDB(dbURL)
	defer models.DB.Close()

	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	r.Use(cors.Default())
	r.Use(gin.Recovery())
	r.Static("/web", "./web")
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Aplikasi Todo")
	})

	srv := services.NewService(models.DB)
	srv.Init(r)
	r.Run()
}
