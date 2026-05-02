package main

import (
	"flag"
	"log"
	"os"

	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkviper "github.com/matiasmartin-labs/common-fwk/config/viper"

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

	// ---- Application ----
	application := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer()

	// ---- Security (v0.9.0: config-based RS256 wiring, exposes GetRSAPrivateKey) ----
	application, err = application.UseServerSecurityFromConfig()
	if err != nil {
		log.Fatalf("wire security from config: %v", err)
	}

	// ---- Health/Readiness presets ----
	if err := application.EnableHealthReadinessPresets(fwkapp.HealthReadinessOptions{}); err != nil {
		log.Fatalf("enable health/readiness: %v", err)
	}

	// ---- Logger ----
	logger, err := application.GetLogger("auth")
	if err != nil {
		log.Fatalf("get logger: %v", err)
	}

	// ---- Routes ----
	if err := server.Routes(application); err != nil {
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
