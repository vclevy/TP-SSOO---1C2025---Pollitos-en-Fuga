package api

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/api/handlers"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"net/http"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.ConfigKernel.Port_Kernel,
		Handlers: map[string]http.HandlerFunc{
			"POST /handshakeCPU":   handlers.HandshakeConCPU,
			"POST /handshakeIO":    handlers.RecibirPaquete,
			"POST /syscallIO":      handlers.IO,
			"POST /finalizacionIO": handlers.FinalizacionIO,
			"POST /Init_Proc":      handlers.INIT_PROC,
			"POST /exit":           handlers.EXIT,
		},
	}
	fmt.Printf("🟢 Kernel prendido en http://localhost:%d\n", global.ConfigKernel.Port_Kernel)
	return server.NuevoServer(configServer)
}
