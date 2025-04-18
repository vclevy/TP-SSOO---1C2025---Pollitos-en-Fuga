package api

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"net/http"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.ConfigMemoria.Port_Memory,
		Handlers: map[string]http.HandlerFunc{
			"POST /escribir": handlers.EscribirMemoria,
		},
	}
	return server.NuevoServer(configServer)
}