package main

import (
	"flag"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"

	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkviper "github.com/matiasmartin-labs/common-fwk/config/viper"

	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/jwks"
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/server"
)

const (
	defaultConfigPath = "./config.yaml"
	configPathEnvKey  = "CONFIG_PATH"
)

func main() {
	// ---- Config ----
	configPath := resolveConfigPath()

	cfg, err := fwkviper.Load(fwkviper.Options{
		ConfigPath: configPath,
		ExpandEnv:  true,
	})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ---- RSA keypair ----
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate rsa key: %v", err)
	}
	keyPair := jwks.NewKeyPair(privateKey)

	// ---- Application ----
	application := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer()

	// ---- Security (v0.4.0: config-based RS256 wiring) ----
	application, err = application.UseServerSecurityFromConfig()
	if err != nil {
		log.Fatalf("wire security from config: %v", err)
	}

	// ---- Health/Readiness presets (v0.6.0) ----
	if err := application.EnableHealthReadinessPresets(fwkapp.HealthReadinessOptions{}); err != nil {
		log.Fatalf("enable health/readiness: %v", err)
	}

	// ---- Logger (v0.7.0) ----
	logger, err := application.GetLogger("auth")
	if err != nil {
		log.Fatalf("get logger: %v", err)
	}

	// ---- Bootstrap ----
	b := &server.Bootstrap{
		PrivateKey: privateKey,
		KeyPair:    keyPair,
		Config:     cfg,
		Logger:     logger,
	}

	if err := server.Routes(application, b); err != nil {
		log.Fatalf("register routes: %v", err)
	}

	logger.Infof("starting auth-provider-ms on port %d", cfg.Server.Port)

	if err := application.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func resolveConfigPath() string {
	configPathFlag := flag.String("config", "", "Path to config file")
	flag.Parse()

	if *configPathFlag != "" {
		return *configPathFlag
	}

	if envPath := os.Getenv(configPathEnvKey); envPath != "" {
		return envPath
	}

	return defaultConfigPath
}
