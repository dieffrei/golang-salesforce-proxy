package main

import (
	"log"
	"github.com/nimajalali/go-force/force"
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"bytes"
	"os"
)

func main() {

	forceApi, err := force.Create(
		os.Getenv("SF_VERSION"),
		os.Getenv("SF_CLIENT_ID"),
		os.Getenv("SF_CLIENT_SECRET"),
		os.Getenv("SF_USER_NAME"),
		os.Getenv("SF_PASSWORD"),
		os.Getenv("SF_TOKEN"),
		os.Getenv("SF_ENVIROMENT"),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("token: " + forceApi.GetAccessToken())
	salesforceSessionId := forceApi.GetAccessToken();

	r := gin.Default()
	r.Static("/bower_components", "./bower_components")
	r.Static("/app", "./app")
	r.LoadHTMLGlob("views/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"sessionId": salesforceSessionId,
		})
	})

	r.Any("/proxy", func (c *gin.Context){
		request := c.Request
		salesforceEndpoint := request.Header.Get("SalesforceProxy-Endpoint")

		client := &http.Client{}
		req, _ := http.NewRequest("GET", salesforceEndpoint, request.Body)
		req.Header.Add("Authorization", request.Header.Get("X-Authorization"))
		req.Header.Add("Content-Type", request.Header.Get("Content-Type"))
		resp, err := client.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		c.String(200, buf.String())
	})
	r.Run(":3004")

}