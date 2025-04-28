package api

import (
	"github.com/sisoputnfrba/tp-golang/kernel/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"net/http"
	"fmt"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.ConfigKernel.Port_Kernel,
		Handlers: map[string]http.HandlerFunc{
			"POST /handshakeCPU" : handlers.HandshakeConCPU,
			"POST /handshakeIO": handlers.RecibirPaquete,
		},
	}
	fmt.Printf("🟢 Kernel prendido en http://localhost:%d\n", global.ConfigKernel.Port_Kernel)
	return server.NuevoServer(configServer)
}