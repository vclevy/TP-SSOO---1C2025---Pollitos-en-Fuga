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
		Port: global.CpuConfig.PortCPU,
		Handlers: map[string]http.HandlerFunc{
			"POST /responder": handlers.RecibirPaqueteDeKernel,
			// http://{ip_kernel}:{port_kernel}/hola
			// ese GET son palabras clave del protocolo http (ver en la docu de go)
		},
	}
	fmt.Printf("🟢 Kernel prendido en http://localhost:%d\n", global.CpuConfig.Port_Kernel)
	return server.NuevoServer(configServer)
}
