package main

import (
	"log"
	"os"

	"github.com/hse-telescope/auth/internal/auth"
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
		log.Default().Printf("\n---CONFIG PARSE ERROR---\n[ERROR]: %s\n", err.Error())
		panic(err)
	}

	if err := auth.InitJWT(conf.JWTSecret); err != nil {
		log.Default().Printf("\n---JWT INIT ERROR---\n[ERROR]: %s\n", err.Error())
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
