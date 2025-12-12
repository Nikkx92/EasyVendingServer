package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

func GetHistory(c Customer) map[string]map[string]int {
	var all = make(map[string]map[string]int)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:sonne@192.168.1.46:5432/postgres?sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT started_at, ended_at, drinks
		FROM customer_request_period
		WHERE inn = $1
		ORDER BY started_at DESC 
`, c.INN)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var start, end time.Time
		var d []string
		if err := rows.Scan(&start, &end, &d); err != nil {
			fmt.Println(err)
		}
		if all[fmt.Sprintf("%s--%s", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))] == nil {
			all[fmt.Sprintf("%s--%s", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))] = make(map[string]int)
			for i := range d {
				all[fmt.Sprintf("%s--%s", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"))][d[i]]++
			}
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return all
}
