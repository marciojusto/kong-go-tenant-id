package main

import (
	"encoding/base64"
	"fmt"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/dgrijalva/jwt-go"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"log"
	"strings"
	"time"
)

var Version = "0.1"
var Priority = 902

type Config struct {
	SecretKeyBase64 string
}

func New() interface{} {
	return &Config{}
}

type Tenant struct {
	TenantResourceUid string    `json:"tenantId"`
	UserResourceUid   string    `json:"userId"`
	ApplicationName   string    `json:"applicationName"`
	RequestURI        string    `json:"requestURI"`
	TimeStamp         time.Time `json:"timeStamp"`
}

var claims = jwt.MapClaims{}

func populateClaims(bearerToken string, config Config) error {
	_, err := jwt.ParseWithClaims(strings.Split(bearerToken, "Bearer ")[1], claims, func(token *jwt.Token) (interface{}, error) {
		decodeString, err := base64.StdEncoding.DecodeString(config.SecretKeyBase64)
		if err != nil {
			log.Fatal("error in decoding secret key")
		}
		return []byte(decodeString), nil
	})
	return err
}

func (config Config) Access(kong *pdk.PDK) {
	bearerToken, err := kong.Request.GetHeader("Authorization")
	if err != nil {
		_ = kong.Log.Alert("error in Get Token from Header")
		_ = kong.Log.Err(err.Error())
	}

	err = populateClaims(bearerToken, config)
	if err != nil {
		_ = kong.Log.Alert("token is invalid or expired")
		_ = kong.Log.Err(err.Error())
	}

	esClient, err := GetESClient(kong)
	if err != nil {
		_ = kong.Log.Alert("error from Elasticsearch")
		_ = kong.Log.Err(err.Error())
	}

	path, err := kong.Request.GetPath()
	if err != nil {
		_ = kong.Log.Alert("error from GetPath")
		_ = kong.Log.Err(err.Error())
	}

	tenantData := Tenant{
		TenantResourceUid: claims["tenantResourceUid"].(string),
		UserResourceUid:   claims["userResourceUid"].(string),
		ApplicationName:   claims["applicationName"].(string),
		RequestURI:        path,
		TimeStamp:         time.Now().Round(0),
	}

	index, err := esClient.Index("tenant", esutil.NewJSONReader(&tenantData))
	if err != nil {
		_ = kong.Log.Alert("error from elasticsearch Index")
		_ = kong.Log.Err(err.Error())
	}
	fmt.Println(index)
}

func GetESClient(kong *pdk.PDK) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	})
	_ = kong.Log.Info("Elasticsearch initialized...")

	return es, err
}

func main() {
	_ = server.StartServer(New, Version, Priority)
}
