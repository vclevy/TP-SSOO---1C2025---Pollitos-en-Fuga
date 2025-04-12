package api

import (
	"github.com/sisoputnfrba/tp-golang/kernel/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"net/http"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.KernelConfig.Port_Kernel,
		Handlers: map[string]http.HandlerFunc{
			"POST /escribir": handlers.EscribirKernel,
			// http://{ip_kernel}:{port_kernel}/hola
			// ese GET son palabras clave del protocolo http (ver en la docu de go)
		},
	}
	return server.NuevoServer(configServer)
}