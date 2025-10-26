package main

import (
	"github.com/gin-gonic/gin"
)

func mainRequest(r Request, c *gin.Context) string {
	codeKit, data := getDataKitVending(r.Data.CompanyId, r.Data.UserLogin, r.Data.PasswordKit, r.Data.Date)
	if codeKit != 0 {
		c.Header("Error-Kit", data)
		return ""
	} else {
		codeFns := check()
		if codeFns == "authentication.failed.expired.token" {
			//codeKit = 1
			//getToken(r.Device, r.INN, r.PasswordFns)
		}
		return codeFns
	}
}
