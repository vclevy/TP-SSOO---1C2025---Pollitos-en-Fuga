package api

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"net/http"
	"fmt"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.ConfigMemoria.Port_Memory,
		Handlers: map[string]http.HandlerFunc{
			"POST /responder": handlers.RecibirPaquete,
			"POST /tamanioProceso": handlers.TamanioProceso,
		},
	}
	fmt.Printf("🟢 Memoria prendida en http://localhost:%d\n", global.ConfigMemoria.Port_Memory)
	return server.NuevoServer(configServer)
}