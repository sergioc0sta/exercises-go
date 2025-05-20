package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
	Api    string
	Cep    string
	City   string
	State  string
	Street string
}

type CepBrasil struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type CepViacep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

const (
	defaultTimeout = 1 * time.Second
	cepToSearch    = "01153000"
)

func getBrasilAPI(ch chan<- Message, ctx context.Context, cep string) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		fmt.Printf("Cant create the request: %v", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Request error: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Status OK: %v", resp.Status)
		return
	}

	defer resp.Body.Close()
	var data CepBrasil
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error to decode to JSON: %v", err)
		return
	}

	select {
	case ch <- Message{Api: "BrasilApi", Cep: data.Cep, City: data.City, State: data.State, Street: data.Street}:
	case <-ctx.Done():
		return
	}
}

func getViacepAPI(ch chan<- Message, ctx context.Context, cep string) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		fmt.Printf("Cant create the request: %v", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Request error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Status OK: %v", resp.Status)
		return
	}

	var data CepViacep
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error to decode to JSON: %v", err)
		return
	}

	select {
	case ch <- Message{Api: "Viacep", Cep: data.Cep, City: data.Localidade, State: data.Estado, Street: data.Bairro}:
	case <-ctx.Done():
		return
	}
}

func main() {
	ch1 := make(chan Message, 1)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	go getBrasilAPI(ch1, ctx, cepToSearch)
	go getViacepAPI(ch1, ctx, cepToSearch)

	select {
	case ms := <-ch1:
		cancel()
		fmt.Printf("API: %s\n", ms.Api)
		fmt.Printf("Cep: %s\n", ms.Cep)
		fmt.Printf("City: %s\n", ms.City)
		fmt.Printf("State: %s\n", ms.State)
		fmt.Printf("Street: %s\n", ms.Street)
	case <-ctx.Done():
		cancel()
		fmt.Printf("Cant get the response! \n")

	}

}
