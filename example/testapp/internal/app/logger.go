//goro:init logger
package app

import (
	"log"
	"testapp/pkg/logger"
)

func (a *App) initLogger() {
	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	a.logger = l
}
