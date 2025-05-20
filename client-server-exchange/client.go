package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	client := &http.Client{}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)

	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	response, err := client.Do(request)

	if err != nil {
		log.Fatal("Erro with request", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)

	log.Println("The response was: ", string(data))

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString("Dólar: {" + string(data)+"}")
	if err != nil {
		log.Fatal("We can't safe the value on file", err)
	}

	log.Println("Value save")
	defer f.Close()
}
