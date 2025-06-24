package main

import (
	"fmt"
	"os"
	"strconv"
	"github.com/sisoputnfrba/tp-golang/kernel/api"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {

	archivoConfig := os.Args[1]
	global.InitGlobal(archivoConfig)
	defer global.LoggerKernel.CloseLogger()

	if len(os.Args) != 4 {
		fmt.Println("Uso: ./kernel <archivo_config> <archivo_pseudocodigo> <tamaño_memoria>")
		os.Exit(1)
	}
	// para pasar como parametro el archivo de config sería config/config.json (o el nombre que le pongamos en las pruebas)

	archivo := os.Args[2]
	tamMemoriaString := os.Args[3]

	tamMemoria, err := strconv.Atoi(tamMemoriaString)
	if err != nil {
		panic(fmt.Sprintf("Tamaño de memoria inválido: %s", tamMemoriaString))
	}

	planificacion.CrearProceso(tamMemoria, archivo)

	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerKernel.Log("Error al iniciar el servidor: "+err_server.Error(), logger.ERROR)
		}
	}()

	go planificacion.IniciarPlanificadorLargoPlazo()

	global.LoggerKernel.Log("Planificador de Largo Plazo en STOP. Presione Enter para iniciar", logger.DEBUG)
	fmt.Scanln()
	close(global.InicioPlanificacionLargoPlazo)

	go planificacion.IniciarPlanificadorCortoPlazo()
	go planificacion.IniciarPlanificadorMedioPlazo()

	select {}
}
