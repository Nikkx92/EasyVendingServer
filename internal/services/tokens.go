package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		fmt.Println(err)
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/token"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var res TokensResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		fmt.Println(err)
	}

	if res.Code == "authentication.failed" {
		return res.Code
	} else {
		return res.Token
	}
}

/*func getRefreshToken(d DeviceInfo, inn, passF string) (string, string) {
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}
	defer resp.Body.Close()

	var res TokensResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		logger.Println(err, "getRefreshToken")
	}

	if res.Code == "authentication.failed" {
		return res.Message, ""
	} else {
		return res.RefreshToken, res.Token
	}

}*/

func getRefreshToken(d DeviceInfo, inn, passF string) (string, string) {
	payload := RefreshToken{
		DeviceInfo: d,
		Username:   inn,
		Password:   passF,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
	}

	url := "https://lknpd.nalog.ru/api/v1/auth/lkfl"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var res TokensResponse
	if err = json.Unmarshal(body, &res); err != nil {
		fmt.Println(err)
	}

	if res.Code == "authentication.failed" {
		return res.Message, ""
	} else {
		return res.RefreshToken, res.Token
	}

}
