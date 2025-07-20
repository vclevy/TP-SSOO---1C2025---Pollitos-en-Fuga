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
			"POST /finalizacionIO": handlers.FinalizacionIO,
			"POST /Init_Proc":      handlers.INIT_PROC,
			"POST /devolucion":     handlers.DevolucionCPUHandler,
		},
	}
	fmt.Printf("🟢 Kernel prendido en http://%s:%d", global.ConfigKernel.Ip_Kernel, global.ConfigKernel.Port_Kernel)
	return server.NuevoServer(configServer)
}
