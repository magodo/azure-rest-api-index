package main

import (
	"log"

	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "azure-rest-api-index",
		Level: hclog.LevelFromString("DEBUG"),
		Color: hclog.AutoColor,
	})

	SetLogger(logger)

	if _, err := BuildIndex("../azure-rest-api-specs/specification", "dedup.json"); err != nil {
		log.Fatal(err)
	}
}
