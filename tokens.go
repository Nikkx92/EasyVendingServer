package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type RefreshToken struct {
	DeviceInfo DeviceInfo `json:"deviceInfo"`
	Username   string     `json:"username"`
	Password   string     `json:"password"`
}

type TokensResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	RefreshToken string `json:"refreshToken"`
	Token        string `json:"token"`
}

type Token struct {
	DeviceInfo   DeviceInfo `json:"deviceInfo"`
	RefreshToken string     `json:"refreshToken"`
}

func getToken(d DeviceInfo, rf string) string {
	payload := Token{
		DeviceInfo:   d,
		RefreshToken: rf,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Println(err, "getToken")
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/token"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Println(err, "getToken")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var response TokensResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Println(err, "getToken")
	}

	if response.Code == "authentication.failed" {
		return response.Code
	} else {
		return response.Token
	}
}

func getRefreshToken(d DeviceInfo, inn, passF string) (string, string, bool) {
	payload := RefreshToken{
		DeviceInfo: d,
		Username:   inn,
		Password:   passF,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/lkfl"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}
	defer resp.Body.Close()

	var response TokensResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}

	fmt.Println(response.Code, response.Message)

	if response.Code == "authentication.failed" {
		return response.Message, "", false
	} else {
		return response.RefreshToken, response.Token, true
	}

}
