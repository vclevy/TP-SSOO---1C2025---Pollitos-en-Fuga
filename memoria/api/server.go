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
			"POST /procesoAMemoria": handlers.InicializarProceso,
			"POST /verificarEspacioDisponible": handlers.VerificarEspacioDisponible,
			"POST /solicitudInstruccion": handlers.DevolverInstruccion,
			"POST /configuracionMMU": handlers.ArmarPaqueteConfigMMU,
			
		},
	}
	fmt.Printf("🟢 Memoria prendida en http://localhost:%d\n", global.ConfigMemoria.Port_Memory)
	return server.NuevoServer(configServer)
}