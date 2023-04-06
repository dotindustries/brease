package main

import (
	"log"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

func tracer() (logger *zap.Logger, client *axiom.Client, flushFn func()) {
	flushFn = func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Fatal(syncErr)
		}
	}

	axiomToken := getenv("AXIOM_TOKEN", "")
	axiomOrg := getenv("AXIOM_ORG", "")

	if axiomToken == "" || axiomOrg == "" {
		logger, _ = zap.NewDevelopment()
		flushFn = func() {
			if syncErr := logger.Sync(); syncErr != nil {
				log.Fatal(syncErr)
			}
		}
		return
	}

	client, err := axiom.NewClient(
		axiom.SetPersonalTokenConfig(axiomToken, axiomOrg),
	)
	if err != nil {
		log.Fatal(err)
	}

	datasetName := getenv("AXIOM_DATASET", "")
	if datasetName != "" {
		core, err := adapter.New(
			adapter.SetClient(client),
			adapter.SetDataset(datasetName),
		)

		if err != nil {
			log.Fatal(err)
		}
		// 2. Spawn the logger.
		logger = zap.New(core, zap.Development())
	}

	logger.Debug("Configured axiom logging", zap.String("dataset", datasetName), zap.String("org", axiomOrg))
	return
}
