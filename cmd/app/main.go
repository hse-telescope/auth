package main

import (
	"os"

	"github.com/hse-telescope/auth/internal/config"
	"github.com/hse-telescope/auth/internal/providers/users"
	"github.com/hse-telescope/auth/internal/repository/facade"
	"github.com/hse-telescope/auth/internal/repository/storage"
	"github.com/hse-telescope/auth/internal/server"
)

func main() {
	configPath := os.Args[1]
	conf, err := config.Parse(configPath)
	if err != nil {
		panic(err)
	}

	storage, err := storage.New(conf.DB.GetDBURL(), conf.DB.MigrationsPath)
	if err != nil {
		panic(err)
	}

	facade := facade.New(storage)

	provider := users.New(facade)

	s := server.New(conf, provider)
	panic(s.Start())
}
