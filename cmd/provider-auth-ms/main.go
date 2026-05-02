package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
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
	// NOTE: we generate the keypair here (not via UseServerSecurityFromConfig) because
	// the app needs direct access to the private key to sign tokens and the public key
	// to expose it via the JWKS endpoint. common-fwk v0.4.0 does not expose the
	// internally generated RSA keypair when using rs256-key-source: generated.
	// See: https://github.com/matiasmartin-labs/common-fwk/issues/50
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

	// ---- Application ----
	application := fwkapp.NewApplication().
		UseConfig(cfg).
		UseServer().
		UseServerSecurity(validator)

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
