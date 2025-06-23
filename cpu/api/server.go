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
			"POST /interrupt": handlers.Interrupcion,
			"POST /dispatch": handlers.NuevoPCB,
 		},
	}
	fmt.Printf("\033[32m🟢 CPU prendido en http://%s:%d\n\033[0m",global.CpuConfig.Ip_Cpu, global.CpuConfig.Port_Cpu)
	return server.NuevoServer(configServer)
}