package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/api"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/kernel/planificacion"
)

func main() {
	// 1. Cargar config y configurar logger
	global.InitGlobal()
	defer global.LoggerKernel.CloseLogger()

	if len(os.Args) != 3 { // go run kernel.go hola.txt 1024
		fmt.Println("Uso: ./kernel <archivo_pseudocodigo> <tamaño_memoria>")
		os.Exit(1)
	}

	archivo := os.Args[1]
	tamMemoriaString := os.Args[2]

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

	// 3. Iniciar la planificación de largo plazo (esperando que se libere)
	planificacion.IniciarPlanificadorLargoPlazo()

	// 4. Espera el ingreso de Enter para liberar la planificación
	fmt.Println("Planificador de Largo Plazo en STOP. Presione Enter para iniciar...")
	fmt.Scanln()  // Bloquea hasta que el usuario presione Enter
	close(global.InicioPlanificacionLargoPlazo)  // Liberamos la planificación

	go planificacion.IniciarPlanificadorMedioPlazo()
	go planificacion.IniciarPlanificadorCortoPlazo()

	select {}
}
