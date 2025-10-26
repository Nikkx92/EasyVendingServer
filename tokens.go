package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	refreshToken string
	token        string
)

type RefreshToken struct {
	DeviceInfo DeviceInfo `json:"deviceInfo"`
	Username   string     `json:"username"`
	Password   string     `json:"password"`
}

type Token struct {
	DeviceInfo   DeviceInfo `json:"deviceInfo"`
	RefreshToken string     `json:"refreshToken"`
}

type RefreshTokenResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	RefreshToken string `json:"refreshToken"`
}

func getToken(d DeviceInfo, inn, passF string) {
	payload := Token{
		DeviceInfo:   d,
		RefreshToken: refreshToken, //"eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJ7XCJ0eXBlXCI6XCJSRUZSRVNIX1RPS0VOXCIsXCJyZWZyZXNoQ29udGV4dFwiOntcImF1dGhUeXBlXCI6XCJMS0ZMXCIsXCJmaWRcIjoxMDAxNzcwMTI5MTEsXCJsb2dpblwiOlwiMTAwMTc3MDEyOTExXCIsXCJpZFwiOjI2NTg2MjAzLFwiZGV2aWNlSWRcIjpcIkZHamRqNzNKaGQtampkRF9kODYzMlwiLFwib3BlcmF0b3JJZFwiOm51bGwsXCJjc3VkVXNlcm5hbWVcIjpudWxsLFwic2VnbWVudHNcIjpbXSxcInRva2VuSWRcIjpudWxsLFwiZmlyc3RTZWdtZW50XCI6bnVsbH0sXCJleHBpcmF0aW9uXCI6bnVsbH0ifQ.8K1UHxEwtsMzX2tMtCIIevUqbG7XDvhR2UAYDCI9pPsNtX21FzpTcIn4VLFzjjwAkzS_Lz1txSlnGweXjMtpOg",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Ошибка при маршалинге JSON: %v", err)
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/token"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func getRefreshToken(d DeviceInfo, inn, passF string) (string, bool) {
	payload := RefreshToken{
		DeviceInfo: d,
		Username:   inn,
		Password:   passF,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Ошибка при маршалинге JSON: %v", err)
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/lkfl"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Ошибка при создании запроса: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	var response RefreshTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Ошибка:", err)
	}

	fmt.Println(response.Code, response.Message)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка при чтении ответа: %v", err)
	}
	fmt.Println(string(body))

	if response.Code == "authentication.failed" {
		return response.Message, false
	} else {
		return response.RefreshToken, true
	}

}
