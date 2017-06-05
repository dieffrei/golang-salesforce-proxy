package main

import (
	"log"
	"github.com/nimajalali/go-force/force"
	"fmt"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"bytes"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Settings struct {
	ApiVersion string `yaml:"sf_version"`
	ClientId string `yaml:"sf_client_id"`
	ClientSecret string `yaml:"sf_client_secret"`
	UserName string `yaml:"sf_user_name"`
	Password string `yaml:"sf_password"`
	Token string `yaml:"sf_token"`
	Enviroment string `yaml:"sf_enviroment"`
	TemplateDir string `yaml:"template_dir"`
	Statics  []string `yaml:"static"`
	Routes  []string `yaml:"routes"`
	ServerPort  string `yaml:"server_port"`
}

func getSettings() (settings *Settings) {
	yamlFileBytes, err := ioutil.ReadFile("sfproxy-settings.yaml")
	yaml.Unmarshal(yamlFileBytes, &settings)
	if err != nil {
		panic(err)
	}
	return settings
}

func getSalesforceConection() (*force.ForceApi, error) {
	settings := getSettings()
	log.Println(settings)
	forceApi, err := force.Create(
		settings.ApiVersion,
		settings.ClientId,
		settings.ClientSecret,
		settings.UserName,
		settings.Password,
		settings.Token,
		settings.Enviroment,
	)
	return forceApi, err
}

func setupStaticDirectories(settings *Settings, router *gin.Engine) {
	for i := 0; i < len(settings.Statics); i++ {
		router.Static("/" + settings.Statics[i], "./" + settings.Statics[i])
	}
}

func setupRoutes(settings *Settings, router *gin.Engine, salesforceSessionId string) {
	if settings.TemplateDir != "" {
		router.LoadHTMLGlob(settings.TemplateDir + "/*")
		for i := 0; i < len(settings.Routes); i++ {
			routeName := settings.Routes[i]
			router.GET("/", func(c *gin.Context) {
				c.HTML(http.StatusOK, routeName + ".tmpl", gin.H{
					"sessionId": salesforceSessionId,
				})
			})
		}
	}
}

func setupRouter(salesforceSessionId string){

	settings := getSettings()
	router := gin.Default()
	setupRoutes(settings, router, salesforceSessionId)

	router.Any("/proxy", func (c *gin.Context){
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
	router.Run(":" + settings.ServerPort)
}

func main() {
	forceApi, error := getSalesforceConection()
	if error != nil {
		panic(error)
	}
	fmt.Println("token: " + forceApi.GetAccessToken())
	salesforceSessionId := forceApi.GetAccessToken();
	setupRouter(salesforceSessionId)
}