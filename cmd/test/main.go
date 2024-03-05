package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type client struct {
	clients.Clients
}

func main() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
	cfg := clients.Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	clients, err := clients.NewClients(cfg)
	check(err)

	client := &client{Clients: clients}

	http.HandleFunc("/", client.handleUpdate)
	http.HandleFunc("/splitwise", client.handleSplitwiseCallback)

	zap.S().Infow("Starting server", "port", 8080)

	zap.S().Fatalw("Server failed", zap.Error(http.ListenAndServe("0.0.0.0:8080", nil)))
}

func (c *client) handleSplitwiseCallback(response http.ResponseWriter, req *http.Request) {
	zap.S().Debug("Handling splitwise callback")
	values := req.URL.Query()
	code := values.Get("code")
	state := values.Get("state")
	err := c.Telegram().Auth(req.Context(), code, state)
	if err != nil {
		zap.S().Errorw("error handling splitwise callback", zap.Error(err))
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Success!"))
}

func (c *client) handleUpdate(response http.ResponseWriter, req *http.Request) {
	zap.S().Debug("Handling update")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		zap.S().Errorw("Failed to read request body", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	update := &gotgbot.Update{}
	err = json.Unmarshal(body, update)
	if err != nil {
		zap.S().Errorw("error parsing update", zap.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Telegram().HandleUpdate(req.Context(), update)
}
