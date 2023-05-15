package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "azure-rest-api-index",
		Level: hclog.LevelFromString("DEBUG"),
		Color: hclog.AutoColor,
	})

	SetLogger(logger)

	index, err := BuildIndex("../azure-rest-api-specs/specification", "dedup.json")
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("/tmp/index.json", b, 0644); err != nil {
		log.Fatal(err)
	}

	var index2 Index
	if err := json.Unmarshal(b, &index2); err != nil {
		log.Fatal(err)
	}

	b2, err := json.MarshalIndent(index2, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if string(b) != string(b2) {
		log.Fatal("unmarshal issue")
	}
}
