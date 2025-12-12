package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

type ResponseKit struct {
	Data []string `json:"data"`
	Buf  string   `json:"buffer"`
}

type ResponseFns struct {
	Message string `json:"Message"`
	Buf     string `json:"buffer"`
}

func checkNewSale(drinks string, c Customer) []string {
	oldData := separateData(drinks)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:sonne@192.168.1.46:5432/postgres?sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	tStart, tEnd := parseTime(c.Date)
	var existsStartDate bool
	err = conn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM customer_request_period
				WHERE inn = $1
				AND started_at = $2
)
`, c.INN, tStart).Scan(&existsStartDate)
	if err != nil {
		fmt.Println(err)
	}

	if existsStartDate {
		var currentDrinks []string
		err := conn.QueryRow(ctx, `
			SELECT drinks
			FROM customer_request_period
			WHERE inn =$1
			AND started_at = $2
	`, c.INN, tStart).Scan(&currentDrinks)
		if err != nil {
			fmt.Println(err)
		}

		if len(oldData) > len(currentDrinks) {
			var end time.Time
			err := conn.QueryRow(ctx, `
				SELECT ended_at
				FROM customer_request_period
				WHERE inn = $1
				AND started_at = $2
`, c.INN, tStart).Scan(&end)
			if err != nil {
				fmt.Println(err)
			}
			c.Date = end.Format("2006-01-02 15:04:05") + "--" + tEnd.Format("2006-01-02 15:04:05")
			_, newData := getDataKitVending(c.CompanyId, c.UserLogin, c.PasswordKit, c.Date)
			drinksAndPrice := separateData(newData)
			return drinksAndPrice
		} else {
			return nil
		}
	} else {
		/*var d []string
		for i := range oldData {
			sep := strings.Split(oldData[i], ":")
			d = append(d, sep[0])
		}

		_, err = conn.Exec(ctx, `
				INSERT INTO customer_request_period (inn, started_at, ended_at, drinks)
				VALUES ($1,$2,$3,$4)
		`, c.INN, tStart, tEnd, d)
		if err != nil {
			fmt.Println("customers_data: ", err)
		}*/
		return oldData
	}
}

func ToKitVending(c Customer) []string {
	codeKit, data := getDataKitVending(c.CompanyId, c.UserLogin, c.PasswordKit, c.Date)
	if codeKit != 0 {
		return nil
	} else {
		newData := checkNewSale(data, c)
		return newData
	}
}

func isDateExist(s string) string {

	/*sepDate := strings.Split(s, "-")
	startDate := parseTime(sepDate[0])
	endDate := parseTime(sepDate[1])
	for k, _ := range dataMap {
		sepDateData := strings.Split(k, "-")
		startDateData := parseTime(sepDateData[0])
		endDateData := parseTime(sepDateData[1])
		if startDate >= startDateData && endDate <= endDateData {
			fmt.Println("за выбранную дату чеки уже выдавались")
			return nil
		} else if startDate >= startDateData && endDate > endDateData {
			fmt.Println("second")
			dateMass = append(dateMass, sepDateData[1]+"-"+sepDate[1])
			return dateMass
			//return sepDateData[1] + "-" + sepDate[1]
		} else if startDate < startDateData && endDate <= endDateData {
			fmt.Println("third")
			dateMass = append(dateMass, sepDate[0]+"-"+sepDateData[0])
		} else if startDate < startDateData && endDate > endDateData {
			//var mass []string
			fmt.Println("four")
			firstDate := sepDate[0] + "-" + sepDateData[0]
			secondDate := sepDateData[1] + "-" + sepDate[1]
			dateMass = append(dateMass, firstDate)
			dateMass = append(dateMass, secondDate)
			return dateMass
		} else {
			dateMass = append(dateMass, sepDate[0]+"-"+sepDate[1])
			return dateMass
		}
	}*/

	return s
}

func convertToDate(i int64) time.Time {
	_, off := time.Now().Zone()
	return time.Unix(i-int64(off), 0)
}

/*func parseTime(s string) int64 {
	const lay = "02.01.06 15:04:05" // DD.MM.YY
	t, err := time.Parse(lay, s)
	if err != nil {
		logger.Println(err)
	}
	sec := t.Unix()
	return sec
}*/

/*func toFns(r Request, c *gin.Context) (string, Request) {
/*var drinks []string
var price []string
var req Request
var er string
if v, found := ca.Get("drinks"); found {
	drinks = v.([]string)
}

if v, found := ca.Get("price"); found {
	price = v.([]string)
}
if v, found := ca.Get("request"); found {
	req = v.(Request)
}*/

/*er, modifiedReq := check(r, c)

	r.Data.RefreshToken = "new refToken"
	r.Data.Token = "new token"
	return "Отправлено успешно", r
}*/

func separateData(s string) []string {
	item := strings.Split(s, ";")
	/*var drink []string
	var price []string
	for i := 0; i < len(item)-1; i++ {
		d := strings.Split(item[i], ":")
		drink = append(drink, d[0])
		price = append(price, d[1])
	}*/
	var data []string
	for i := range len(item) - 1 {
		data = append(data, item[i])
	}

	return data
}
