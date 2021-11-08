package main

import (
	"encoding/base64"
	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/dgrijalva/jwt-go"
	"log"
	"strings"
)

var Version = "0.2"
var Priority = 902

func main() {
	err := server.StartServer(New, Version, Priority)
	if err != nil {
		log.Fatal("error from StartServer")
	}
}

type Config struct {
	SecretKeyBase64 string
	HeaderName      string
}

func New() interface{} {
	return &Config{}
}

func (config Config) Access(kong *pdk.PDK) {
	bearerToken, err := kong.Request.GetHeader("Authorization")
	if err != nil {
		log.Fatal("error from GetHeader")
	}

	tenantId, err := fetchTenantId(bearerToken, config)
	if err != nil {
		log.Fatalln("token is invalid or expired")
	}

	err = kong.Response.SetHeader(config.HeaderName, tenantId)
	if err != nil {
		log.Fatal("error from SetHeader")
	}
}

func fetchTenantId(bearerToken string, config Config) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(strings.Split(bearerToken, "Bearer ")[1], claims, func(token *jwt.Token) (interface{}, error) {
		decodeString, err := base64.StdEncoding.DecodeString(config.SecretKeyBase64)
		if err != nil {
			log.Fatal("error in decoding secret key")
		}
		return []byte(decodeString), nil
	})
	return claims["tenantId"].(string), err
}
