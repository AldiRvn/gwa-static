package config

import (
	"log"
	"os"
)

const (
	GENERATED_FOLDER = "./gen"
	DB_FOLDER        = "./db"
)

func init() {
	if err := os.MkdirAll(GENERATED_FOLDER, os.ModePerm); err != nil {
		log.Println(err)
	}
}
