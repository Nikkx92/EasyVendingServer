package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type Request struct {
	Data    *LoginData  `json:"Data"`
	Device  *DeviceInfo `json:"Device"`
	DataKit *[]string   `json:"dataKit"`
}

type MetaDetails struct {
	UserAgent string `json:"userAgent"`
}

type DeviceInfo struct {
	SourceDeviceID string       `json:"sourceDeviceId"`
	SourceType     string       `json:"sourceType"`
	AppVersion     string       `json:"appVersion"`
	MetaDetails    *MetaDetails `json:"metaDetails"`
}

type LoginData struct {
	CompanyId   string `json:"CompanyId"`
	UserLogin   string `json:"UserLogin"`
	PasswordKit string `json:"PasswordKit"`
	INN         string `json:"INN"`
	PasswordFns string `json:"PasswordFns"`
}

type AuthResponse struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const prefix = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "missing bearer token"})
			fmt.Println("missing bearer token")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, prefix)
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid token"})
			fmt.Println(err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid token"})
			fmt.Println("invalid token claims")
			return
		}

		sub, ok := claims["sub"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// после декодирования JWT число обычно становится float64
		subFloat, ok := sub.(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid sub type"})
			return
		}
		userId := int64(subFloat)
		c.Set("jwt", tokenString)
		c.Set("userId", userId)
		c.Next()
	}
}

func Authorization(r Request) (bool, string) {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var c Customer
	var idCustomer int64
	err = conn.QueryRow(ctx, `
		SELECT
			id, inn, company_id, user_login, password_kit, password_fns
		FROM customers_data
		WHERE inn = $1
			AND company_id = $2
			AND user_login = $3
`, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin).Scan(&idCustomer, &c.INN, &c.CompanyId, &c.UserLogin, &c.PasswordKit, &c.PasswordFns)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("customers_data: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}

	var token string
	err = conn.QueryRow(ctx, `
		SELECT 
		    jwt, device_id
		FROM customers_device
		WHERE customers_data_id = $1
`, idCustomer).Scan(&token, &c.Device.SourceDeviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("customers_device: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}

	ok :=
		r.Data.INN == c.INN &&
			r.Data.CompanyId == c.CompanyId &&
			r.Data.UserLogin == c.UserLogin &&
			r.Data.PasswordKit == c.PasswordKit &&
			r.Data.PasswordFns == c.PasswordFns &&
			r.Device.SourceDeviceID == c.Device.SourceDeviceID

	if ok {
		return true, token
	} else {
		return false, ""
	}
}
