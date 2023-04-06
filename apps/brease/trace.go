package main

import (
	"log"
	"os"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func tracer() (logger *zap.Logger, client *axiom.Client, flushFn func()) {
	flushFn = func() {
		if syncErr := logger.Sync(); syncErr != nil {
			log.Fatal(syncErr)
		}
	}

	pe := zap.NewDevelopmentEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEnc := zapcore.NewConsoleEncoder(pe)
	consoleCore := zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), zap.DebugLevel)

	axiomToken := getenv("AXIOM_TOKEN", "")
	axiomOrg := getenv("AXIOM_ORG", "")

	if axiomToken == "" || axiomOrg == "" {
		logger = zap.New(consoleCore)
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
		core = zapcore.NewTee(core, consoleCore)
		logger = zap.New(core, zap.Development())
	}

	return
}
