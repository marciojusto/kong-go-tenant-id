package main

import (
	"encoding/base64"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strings"
)

var Version = "0.3"
var Priority = 903

type Config struct {
	SecretKeyBase64 string
}

func New() interface{} {
	return &Config{}
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
		log.Fatal("error from GetHeader")
	}

	err = populateClaims(bearerToken, config)
	if err != nil {
		log.Fatalln("token is invalid or expired")
	}

	headers := map[string][]string{
		"X-Kong-Tenant-Id":          {claims["tenantResourceUid"].(string)},
		"X-Kong-User-Id":            {claims["userResourceUid"].(string)},
		"X-Kong-ApplicationName-Id": {claims["applicationName"].(string)},
	}
	err = kong.Response.SetHeaders(headers)
	if err != nil {
		log.Fatal("error from SetHeaders")
	}
}

func main() {
	err := server.StartServer(New, Version, Priority)
	if err != nil {
		log.Fatal("error from StartServer")
	}
}
