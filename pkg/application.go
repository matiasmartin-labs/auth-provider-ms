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
	handler *gin.Engine
	KeyPair KeyPair
}

var App *Application

func NewApplication() *Application {
	App = &Application{}
	return App
}

func (app *Application) UseConfig() *Application {
	app.Config = NewConfiguration()
	return app
}

func (app *Application) UseServer() *Application {
	serverProps := app.Config.GetServerConfig()
	r := gin.Default()
	app.handler = r

	app.Server = http.Server{
		Addr:           fmt.Sprintf(":%d", serverProps.GetPort()),
		Handler:        r,
		ReadTimeout:    serverProps.GetReadTimeout(),
		WriteTimeout:   serverProps.GetWriteTimeout(),
		MaxHeaderBytes: serverProps.GetMaxHeaderBytes(),
	}
	return app
}

func (app *Application) RegisterGET(path string, handler gin.HandlerFunc) {
	app.handler.GET(path, handler)
}

func (app *Application) RegisterProtectedGET(path string, handler gin.HandlerFunc) {
	app.handler.GET(path, app.AuthMiddleware(), handler)
}

func (app *Application) Run(entryPoint EntryPointFunc) error {
	err := entryPoint(app)
	if err != nil {
		return err
	}

	return app.Server.ListenAndServe()
}
