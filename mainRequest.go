package main

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func mainRequest(r Request, c *gin.Context) string {
	codeKit, data := getDataKitVending(r.Data.CompanyId, r.Data.UserLogin, r.Data.PasswordKit, r.Data.Date)
	if codeKit != 0 {
		c.Header("Error-Kit", data)
		return ""
	} else {
		d, p := separateData(data)
		const maxRetries = 1
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if r.Data.RefreshToken == "" {
				rt, t, ok := getRefreshToken(r.Device, r.Data.INN, r.Data.PasswordFns)
				if ok {
					c.Header("Refresh-Token", rt)
					c.Header("Token", t)
					r.Data.RefreshToken = rt
					r.Data.Token = t
				} else {
					c.Header("Invalid-Refresh-Token", rt)
				}
			}

			er := check(r, d, p)
			if er == "authentication.failed" && attempt < maxRetries {
				r.Data.RefreshToken = ""
				r.Data.Token = ""
				continue
			}
		}
		return ""
	}
}

func separateData(s string) ([]string, []string) {
	item := strings.Split(s, ";")
	var drink []string
	var price []string
	for i := 0; i < len(item)-1; i++ {
		d := strings.Split(item[i], ":")
		drink = append(drink, d[0])
		price = append(price, d[1])
	}
	return drink, price
}
