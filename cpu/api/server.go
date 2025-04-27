package api
import (
	"github.com/sisoputnfrba/tp-golang/cpu/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"net/http"
	"fmt"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.CpuConfig.Port_Cpu,
		Handlers: map[string]http.HandlerFunc{
		/* "POST /responder": handlers.RecibirPaquete, */
			"POST /handshake": handlers.HandshakeConKernel,
		},
	}
	fmt.Printf("🟢 CPU prendido en http://localhost:%d\n", global.CpuConfig.Port_Cpu)
	return server.NuevoServer(configServer)
}