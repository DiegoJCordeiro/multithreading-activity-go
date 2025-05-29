package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Endereco struct {
	Cep          string `json:"cep"`
	Logradouro   string `json:"logradouro"`
	Bairro       string `json:"bairro"`
	Localidade   string `json:"localidade"` // cidade
	Uf           string `json:"uf"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"` // cidade
	State        string `json:"state"`
	Origem       string // BrasilAPI ou ViaCEP
}

func main() {
	cep := "01153000"

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resultChan := make(chan Endereco, 1)

	go buscarBrasilAPI(ctx, cep, resultChan)
	go buscarViaCEP(ctx, cep, resultChan)

	select {
	case <-ctx.Done():
		log.Println("Timeout: Nenhuma API respondeu em 1 segundo.")
	case endereco := <-resultChan:
		showInfo(endereco)
	}
}

func showInfo(endereco Endereco) {
	fmt.Printf("Endereço encontrado pela API %s:\n", endereco.Origem)
	if endereco.Origem == "BrasilAPI" {
		fmt.Printf("CEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nUF: %s\n",
			endereco.Cep, endereco.Street, endereco.Neighborhood, endereco.City, endereco.State)
	} else {
		fmt.Printf("CEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nUF: %s\n",
			endereco.Cep, endereco.Logradouro, endereco.Bairro, endereco.Localidade, endereco.Uf)
	}
}

func buscarBrasilAPI(ctx context.Context, cep string, resultChan chan Endereco) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var endereco Endereco
	if err := json.NewDecoder(resp.Body).Decode(&endereco); err != nil {
		return
	}
	endereco.Origem = "BrasilAPI"

	select {
	case resultChan <- endereco:
	case <-ctx.Done():
	}
}

func buscarViaCEP(ctx context.Context, cep string, resultChan chan Endereco) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var endereco Endereco
	if err := json.NewDecoder(resp.Body).Decode(&endereco); err != nil {
		return
	}
	endereco.Origem = "ViaCEP"

	select {
	case resultChan <- endereco:
	case <-ctx.Done():
	}
}
