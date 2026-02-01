package main

import (
	"github.com/matiasmartin-labs/auth-provider-ms/internal/infrastructure/port/in/server"
	"github.com/matiasmartin-labs/auth-provider-ms/pkg"
)

func main() {
	pkg.NewApplication().
		UseConfig().
		UseServer().
		Run(server.Routes)
}
