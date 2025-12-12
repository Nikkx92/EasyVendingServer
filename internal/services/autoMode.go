package services

import (
	"context"
	"log"
	"sync"
	"time"
)

type CancelMap struct {
	sync.Mutex
	internal map[string]context.CancelFunc
}

func NewCancelMap() *CancelMap {
	return &CancelMap{
		internal: make(map[string]context.CancelFunc),
	}
}

func (c *CancelMap) Get(key string) (context.CancelFunc, bool) {
	c.Lock()
	defer c.Unlock()
	cancel, ok := c.internal[key]
	return cancel, ok
}

func (c *CancelMap) Set(key string, cancel context.CancelFunc) {
	c.Lock()
	defer c.Unlock()
	c.internal[key] = cancel
}

func (c *CancelMap) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.internal, key)
}

func BackgroundWork(ctx context.Context, cus Customer, id int64) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	//startTime := time.Now().Format("2006-01-02" + " 00:00:00")

	/*work := func() {
		nowTime := time.Now().Format("2006-01-02 15:04:05")
		cus.Date = startTime + "--" + nowTime
		drinks := toKitVending(cus)
		if drinks == nil {
			return
		} else {
			addDrinksDb(cus, drinks)
			message, modifiedCustomer := check(cus, drinks)
			fnsTokenUpdate(modifiedCustomer, id)
			fmt.Println(message)
		}
	}*/

	select {
	case <-ctx.Done():
		log.Printf("Автоматическая отправка для %s остановлена", cus.INN)
		return
	default:
		//work()
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Автоматическая отправка для %s остановлена", cus.INN)
			return
		case <-ticker.C:
			//work()
		}
	}
}
