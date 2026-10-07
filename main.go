package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type printjob struct {
	PrinterHostname string `json:"printerHostname"`
	Text            string `json:"text"`
}

var defaultAllowOrigins = []string{
	"https://www.browser-print.vercel.app",
	"https://browser-print.vercel.app",
	"http://localhost",
	"http://localhost:3000",
	"http://localhost:5173",
	"https://nafcosa.vercel.app",
	"https://www.nafcosa.vercel.app",
	"http://nafcosa.vercel.app",
	"https://nafcosa2.vercel.app",
	"https://www.nafcosa2.vercel.app",
	"http://nafcosa2.vercel.app",
}

func getAllowOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("ALLOW_ORIGINS"))
	if raw == "" {
		origins := make([]string, len(defaultAllowOrigins))
		copy(origins, defaultAllowOrigins)
		return origins
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	if len(origins) == 0 {
		fallback := make([]string, len(defaultAllowOrigins))
		copy(fallback, defaultAllowOrigins)
		return fallback
	}

	return origins
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func handlePrint(c *gin.Context) {
	var newPrintJob printjob
	if err := c.BindJSON(&newPrintJob); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("error parsing JSON: %v", err),
		})
		return
	}

	// send data to printer
	connection, err := net.Dial("tcp", newPrintJob.PrinterHostname+":9100")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error connecting to printer: %v", err),
		})
		return
	}
	defer connection.Close()
	_, err = connection.Write([]byte(newPrintJob.Text))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error sending ZPL to printer: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	r.Use(cors.New(cors.Config{
		AllowOrigins: getAllowOrigins(),
		AllowMethods: []string{"POST"},
		AllowHeaders: []string{"Origin", "Content-Type"}, // TODO: Acess to fetch at 'http://localhost:8080/print' from origin 'http://localhost:3000' has been blocked by CORS policy: Request header field content-type is not allowed by Access-Control-Allow-Headers in preflight response.
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return origin == "https://github.com"
		// },
		// MaxAge: 12 * time.Hour,
	}))
	r.GET("/ping", handlePing)
	r.POST("/print", handlePrint)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
