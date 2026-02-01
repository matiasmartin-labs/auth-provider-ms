package pkg

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EntryPointFunc func(*Application) error

type Application struct {
	Config  Configuration
	Server  http.Server
	Handler *gin.Engine
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) UseConfig() *Application {
	app.Config = NewConfiguration()
	return app
}

func (app *Application) UseServer() *Application {
	serverProps := app.Config.GetServerConfig()
	r := gin.Default()
	app.Handler = r

	app.Server = http.Server{
		Addr:           fmt.Sprintf(":%d", serverProps.GetPort()),
		Handler:        r,
		ReadTimeout:    serverProps.GetReadTimeout(),
		WriteTimeout:   serverProps.GetWriteTimeout(),
		MaxHeaderBytes: serverProps.GetMaxHeaderBytes(),
	}
	return app
}

func (app *Application) Run(entryPoint EntryPointFunc) error {
	err := entryPoint(app)
	if err != nil {
		return err
	}

	return app.Server.ListenAndServe()
}
