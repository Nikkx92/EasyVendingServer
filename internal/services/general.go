package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"strings"
	"time"
)

var (
	Buf         bytes.Buffer
	logger      = log.New(&Buf, "server: ", log.Ltime+log.Lshortfile)
	Jobs        = NewCancelMap()
	jwtSecret   = []byte("Don't Be a Menace to South Central While Drinking Your Juice in the Hood")
	WorkingAuto = make(map[string]int)
)

type SingleRequest struct {
	Date     string   `json:"date"`
	DataKit  []string `json:"dataKit"`
	AutoMode bool     `json:"autoMode"`
	IsPaid   bool     `json:"isPaid"`
}

type Customer struct {
	CompanyId    string     `json:"CompanyId"`
	UserLogin    string     `json:"UserLogin"`
	PasswordKit  string     `json:"PasswordKit"`
	INN          string     `json:"INN"`
	PasswordFns  string     `json:"PasswordFns"`
	Date         string     `json:"Date"`
	Device       DeviceInfo `json:"Device"`
	RefreshToken string     `json:"RefreshToken"`
	Token        string     `json:"TokenType"`
}

func newPGConn() (context.Context, *pgx.Conn, error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx,
		"postgres://postgres:sonne@192.168.1.46:5432/postgres?sslmode=disable",
	)
	if err != nil {
		return nil, nil, err
	}

	return ctx, conn, nil
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

func AddCustomer(r Request, refreshToken, token string) string {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var customersDataID int64
	err = conn.QueryRow(ctx, `
		INSERT INTO customers_data (inn, company_id, user_login, password_kit, password_fns, auto_mode, is_paid)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
`, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin, r.Data.PasswordKit, r.Data.PasswordFns, false, false).Scan(&customersDataID)
	if err != nil {
		fmt.Println("customers_data: ", err)
	}

	jwtToken := newToken(customersDataID)

	_, err = conn.Exec(ctx, `
		INSERT INTO customers_device(
		          jwt,customers_data_id, device_id,refresh_token,token,last_token                   
		)
		VALUES ($1,$2,$3,$4,$5,$6)
`, jwtToken, customersDataID, r.Device.SourceDeviceID, refreshToken, token, time.Now())
	if err != nil {
		fmt.Println("customers_device: ", err)
	}
	return jwtToken
}

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

func InitData(r SingleRequest, id int64) Customer {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var c Customer
	err = conn.QueryRow(ctx, `
		SELECT
			inn, company_id, user_login, password_kit, password_fns
		FROM customers_data
		WHERE id = $1
`, id).Scan(&c.INN, &c.CompanyId, &c.UserLogin, &c.PasswordKit, &c.PasswordFns)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("customers_data: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}

	err = conn.QueryRow(ctx, `
		SELECT
			device_id, refresh_token, token
		FROM customers_device
		WHERE customers_data_id = $1
`, id).Scan(&c.Device.SourceDeviceID, &c.RefreshToken, &c.Token)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("customers_data: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}
	c.Date = r.Date
	return c
}

func AddDrinksDb(c Customer, drinks []string) {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var d []string
	for i := range drinks {
		sep := strings.Split(drinks[i], ":")
		d = append(d, sep[0])
	}
	tStart, tEnd := parseTime(c.Date)

	tag, err := conn.Exec(ctx, `
		INSERT INTO customer_request_period (inn,started_at, ended_at, drinks)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (inn,started_at) DO NOTHING;
`, c.INN, tStart, tEnd, d)
	if err != nil {
		fmt.Println(err)
	}

	if tag.RowsAffected() == 0 {
		_, err = conn.Exec(ctx, `
					UPDATE customer_request_period
					SET
					    drinks = COALESCE(drinks, ARRAY[]::text[]) || $3,
						ended_at = $4
					WHERE inn = $2
					AND started_at = $1
	`, tStart, c.INN, d, tEnd)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func FnsTokenUpdate(c Customer, id int64) {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var lastToken time.Time
	err = conn.QueryRow(ctx, `
					SELECT last_token FROM customers_device WHERE customers_data_id = $1;
		`, id).Scan(&lastToken)
	if errors.Is(err, pgx.ErrNoRows) {
		fmt.Println("customer not exist")
	}
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()
	if lastToken.Before(now.Add(-1 * time.Hour)) {
		_, err = conn.Exec(ctx, `
						UPDATE customers_device
						SET 
						    last_token = $3,
							token = $2
						WHERE customers_data_id = $1;
		`, id, c.Token, now)
	}
	if err != nil {
		fmt.Println(err)
	}
}
