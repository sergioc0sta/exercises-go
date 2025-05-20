package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Exchange struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Usd struct {
	ID        int       `gorm:"primaryKey"`
	Value     float64   `gorm:"column:value"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func getExchangeUsdBrl() (*Exchange, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)

	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var dataParser Exchange

	if err := json.Unmarshal(data, &dataParser); err != nil {
		return nil, err
	}

	return &dataParser, nil
}

func getExchange(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	dataParser, err := getExchangeUsdBrl()

	if err != nil {
		log.Println("Problems to return the value exchange", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	value, err := strconv.ParseFloat(dataParser.USDBRL.Bid, 64)

	if err != nil {
		log.Println("Error converting value:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usd := Usd{
		Value:     value,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	result := db.WithContext(ctx).Create(&usd)
	if result.Error != nil {
		log.Println("Error saving to database:", result.Error)

	}
	fmt.Println("The value of USD-BRL is:", dataParser.USDBRL.Bid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dataParser.USDBRL.Bid))
}

func main() {
	db, err := gorm.Open(sqlite.Open("exchange.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Impossible to open database:", err)
	}

	err = db.AutoMigrate(&Usd{})
	if err != nil {
		log.Fatal("Problems with migrations:", err)
	}

	log.Println("Database is running...")

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		getExchange(db, w, r)
	})
	http.ListenAndServe(":8080", nil)
}
