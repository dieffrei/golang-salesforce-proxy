package main

import (
	"log"
	"github.com/nimajalali/go-force/force"
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"net/http/httputil"
	"net/url"
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

/*	forceApi, err := force.Create(
		"v36.0",
		"3MVG93MGy9V8hF9NrjdwnYswsF4RoGsK0pvcGxiXVZop4LfliEJ59LSgs.uwEIohaAaQ0gMq0LRLAnjUDGFfh",
		"5754176615975639672",
		"dieffrei.quadros@saint-gobain.com.dev3",
		"Topi.1234!@!@",
		"xPCq179X57quwwYWWlkZ5qo9h",
		"sandbox",
	)
*/

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

func ReverseProxy(salesforceSessionId string) gin.HandlerFunc {
	return func(c *gin.Context) {
		director := func(req *http.Request) {
			url, _ := url.Parse(req.Header.Get("SalesforceProxy-Endpoint"))
			fmt.Println("proxy url: " + url.String())
			r := c.Request
			req = r
			req.URL.Scheme = "https"
			req.URL.Host = url.Host
			req.Header["Authorization"] = []string{r.Header.Get("X-Authorization")}
			req.Header["Content-Type"] = []string{r.Header.Get("Content-Type")}
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}