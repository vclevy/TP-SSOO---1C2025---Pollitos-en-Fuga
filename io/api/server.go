package api

import (
	"github.com/sisoputnfrba/tp-golang/io/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/io/global"
	"net/http"
	"fmt"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.IoConfig.Port_Io,
		Handlers: map[string]http.HandlerFunc{
			"POST /responder": handlers.RecibirPaquete,
			"POST /procesoRecibido": handlers.ProcesoRecibidoHandler,
		},
	}
	fmt.Printf(" IO esperando respuesta en http://localhost:%d\n", global.IoConfig.Port_Io)
	return server.NuevoServer(configServer)
}