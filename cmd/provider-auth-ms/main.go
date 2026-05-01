package main

import (
	"flag"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"

	fwkapp "github.com/matiasmartin-labs/common-fwk/app"
	fwkviper "github.com/matiasmartin-labs/common-fwk/config/viper"
	fwkjwt "github.com/matiasmartin-labs/common-fwk/security/jwt"
	fwkkeys "github.com/matiasmartin-labs/common-fwk/security/keys"

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

	// ---- JWT validator ----
	resolver := fwkkeys.NewRSAResolver(privateKey, keyPair.KeyID)
	validator, err := fwkjwt.NewValidator(fwkjwt.Options{
		Methods:  []string{"RS256"},
		Issuer:   cfg.Security.Auth.JWT.Issuer,
		Resolver: resolver,
	})
	if err != nil {
		log.Fatalf("create validator: %v", err)
	}

	// ---- Bootstrap ----
	b := &server.Bootstrap{
		PrivateKey: privateKey,
		KeyPair:    keyPair,
		Config:     cfg,
	}

	// ---- Application ----
	app := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer().
		UseServerSecurity(validator)

	if err := server.Routes(app, b); err != nil {
		log.Fatalf("register routes: %v", err)
	}

	if err := app.Run(); err != nil {
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
