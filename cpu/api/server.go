package api
import (
	"github.com/sisoputnfrba/tp-golang/cpu/api/handlers"
	"github.com/sisoputnfrba/tp-golang/utils/server"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"net/http"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.CpuConfig.PortCPU,
		Handlers: map[string]http.HandlerFunc{
			"POST /responder": handlers.SaludoAKernel,
			// http://{ip_kernel}:{port_kernel}/hola
			// ese GET son palabras clave del protocolo http (ver en la docu de go)
		},
	}
	return server.NuevoServer(configServer)
}