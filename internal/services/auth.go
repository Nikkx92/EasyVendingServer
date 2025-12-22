package services

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"strings"
	"test/internal/domain"
	"test/internal/fns"
	"test/internal/kitvending"
	"test/internal/store"
	"time"
)

type Request struct {
	Data    *domain.LoginData  `json:"Data"`
	Device  *domain.DeviceInfo `json:"Device"`
	DataKit *[]string          `json:"dataKit"`
}

type AuthResponse struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
}

type FNSAuth interface {
	//GetToken(ctx context.Context, d *domain.DeviceInfo, rf string) (string, error)
	GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error)
	//SendSaleToFns(ctx context.Context, cus domain.Customer, drinks []string) (domain.Customer, error)
}

type KitAuth interface {
	ValidLoginKit(ctx context.Context, companyId, userLogin, pass string) error
	//GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (int, string, error)
}

type AuthService struct {
	store   *store.Store
	fnsAuth FNSAuth
	kitAuth KitAuth
}

func NewAuthService(st *store.Store, fnsC *fns.Client, kitC *kitvending.Client) *AuthService {
	return &AuthService{
		store:     st,
		fnsClient: fnsC,
		kitClient: kitC,
	}
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

		userId := GetIdFromJwt(tokenString)

		c.Set("jwt", tokenString)
		c.Set("userId", userId)
		c.Next()
	}
}

func GetIdFromJwt(t string) int64 {
	token, err := jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		slog.Info("invalid token")
		return 0
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Info("invalid token claims")
	}

	sub, ok := claims["sub"]
	if !ok {
		slog.Info("invalid token sub")
	}

	subFloat, ok := sub.(float64)
	if !ok {
		slog.Info("invalid token sub")
	}
	return int64(subFloat)
}

/*
	func Authorization(ctx context.Context, r Request) (string, error) {
		conn, err := newPGConn(ctx)
		if err != nil {
			slog.Error(err.Error(), "ИНН", r.Data.INN)
			return "", fmt.Errorf("сервис бд не доступен, попробуйте позже")
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
			slog.Info("в БД нет такой комбинации", "ИНН", r.Data.INN)
		}
		slog.Info("no rows in result set", "ИНН", r.Data.INN)
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
				slog.Info("в БД нет такой комбинации", "ИНН", r.Data.INN)
			}
			slog.Info("no rows in result set", "ИНН", r.Data.INN)
		}

		ok :=
			r.Data.INN == c.INN &&
				r.Data.CompanyId == c.CompanyId &&
				r.Data.UserLogin == c.UserLogin &&
				r.Data.PasswordKit == c.PasswordKit &&
				r.Data.PasswordFns == c.PasswordFns &&
				r.Device.SourceDeviceID == c.Device.SourceDeviceID

		if ok {
			return token, nil
		} else {
			return "", nil
		}
	}
*/
func newToken(id int64) string {
	claims := jwt.MapClaims{
		"sub": id, // идентификатор пользователя
		//"exp": time.Now().Add(30 * time.Minute).Unix(), // срок действия access‑токена
		//"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		fmt.Println(err)
	}
	return tokenString
}

func parseTime(s string) (time.Time, time.Time) {
	sep := strings.Split(s, "--")
	const lay = "2006-01-02 15:04:05"
	loc := time.Local
	t1, err := time.ParseInLocation(lay, sep[0], loc)
	if err != nil {
		fmt.Println(err)
	}
	t2, err := time.ParseInLocation(lay, sep[1], loc)
	if err != nil {
		fmt.Println(err)
	}
	return t1, t2
}

func (a *AuthService) Authorization(ctx context.Context, r Request) (string, error) {
	cd, err := a.store.GetCustomerDataByInn(ctx, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin)
	if err != nil {
		return "", err
	}

	dev, err := a.store.GetCustomerDeviceId(ctx, cd.IDCustomer)
	if err != nil {
		return "", err
	}

	ok :=
		r.Data.INN == cd.INN &&
			r.Data.CompanyId == cd.CompanyId &&
			r.Data.UserLogin == cd.UserLogin &&
			r.Data.PasswordKit == cd.PasswordKit &&
			r.Data.PasswordFns == cd.PasswordFns &&
			r.Device.SourceDeviceID == dev.DeviceID

	if !ok {
		return "", nil
	}
	return dev.JWT, nil
}

func (a *AuthService) AddCustomer(ctx context.Context, r Request, refreshToken, token string) (string, error) {
	var jwtToken string

	data := store.CustomerData{
		CompanyId:   r.Data.CompanyId,
		UserLogin:   r.Data.UserLogin,
		PasswordKit: r.Data.PasswordKit,
		INN:         r.Data.INN,
		PasswordFns: r.Data.PasswordFns,
		AutoMode:    false,
		IsPaid:      false,
	}

	err := pgx.BeginFunc(ctx, a.store.Pool, func(tx pgx.Tx) error {
		id, err := a.store.InsertCustomerData(ctx, data, tx)
		if err != nil {
			return err
		}

		jwtToken = newToken(id)
		dev := store.CustomerDevice{
			JWT:      jwtToken,
			DeviceID: r.Device.SourceDeviceID,
		}

		if err := a.store.InsertCustomerDevice(ctx, dev, refreshToken, token, tx, id); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return jwtToken, nil
}

/*func (a *AuthService) GetCustomerByID(ctx context.Context, id int64) (domain.Customer, error) {
	var c domain.Customer
	c, err := a.store.GetCustomerDataByID(ctx, id)
	if err != nil {
		return c, err
	}
	return c, nil
}*/

/*func (a *AuthService) CheckNewSale(ctx context.Context, c domain.Customer, drinks, date string) ([]string, error) {
	newDrinks := separateData(drinks)
	tStart, tEnd := parseTime(date)

	exists, err := a.store.ExistsStartDate(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	if !exists {
		return newDrinks, nil
	}

	oldDrinks, err := a.store.GetDrinks(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	if len(newDrinks) <= len(oldDrinks) {
		return nil, nil
	}

	end, err := a.store.GetEndDate(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	date = end.Format("2006-01-02 15:04:05") + "--" + tEnd.Format("2006-01-02 15:04:05")
	_, newData, err := a.kitClient.GetDataKitVending(ctx, c.CompanyID, c.UserLogin, c.PasswordKit, date)
	if err != nil {
		return nil, err
	}
	return separateData(newData), nil
}*/

func (a *AuthService) AddDrinks(ctx context.Context, c domain.Customer, drinks []string, date string) error {
	tStart, tEnd := parseTime(date)
	var titleDrinks []string

	for i := range drinks {
		sep := strings.Split(drinks[i], ":")
		titleDrinks = append(titleDrinks, sep[0])
	}
	if err := a.store.InsertDrinks(ctx, c.INN, titleDrinks, tStart, tEnd); err != nil {
		return err
	}
	return nil
}

/*func (a *AuthService) GetToken(ctx context.Context, d *domain.DeviceInfo, rf string) (string, error) {
	return a.fnsClient.GetToken(ctx, d, rf)
}*/

func (a *AuthService) GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error) {
	return a.fnsAuth.GetRefreshToken(ctx, d, inn, pass)
}

/*func (a *AuthService) SendSaleToFns(ctx context.Context, cus domain.Customer, drinks []string) (domain.Customer, error) {
	return a.fnsClient.SendSaleToFns(ctx, cus, drinks)
}*/
