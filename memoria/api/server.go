package api

import (
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/memoria/api/handlers"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/server"
)

func CrearServer() *server.Server {
	configServer := server.Config{
		Port: global.ConfigMemoria.Port_Memory,
		Handlers: map[string]http.HandlerFunc{
			//usadas por el KERNEL
			"POST /inicializarProceso": handlers.InicializarProceso,
			"POST /verificarEspacioDisponible": handlers.VerificarEspacioDisponible,
			"POST /finalizarProceso": handlers.FinalizarProceso,
			"POST /suspension": handlers.SuspenderProceso,
			"POST /dessuspension": handlers.DesSuspenderProceso,

			//usadas por la CPU
			"POST /solicitudInstruccion": handlers.DevolverInstruccion,
			"POST /configuracionMMU": handlers.ArmarPaqueteConfigMMU,
			"POST /pedirMarco": handlers.AccederTablaPaginas,
			"POST /leerMemoria": handlers.LeerMemoria,
			"POST /escribirMemoria": handlers.EscribirMemoria,
			"POST /leerPaginaCompleta": handlers.LeerPaginaCompleta,
			"POST /actualizarPaginaCompleta": handlers.EscribirPaginaCompleta,
		},
	}
	fmt.Printf("🟢 Memoria prendida en http://localhost:%d\n", global.ConfigMemoria.Port_Memory)
	return server.NuevoServer(configServer)
}