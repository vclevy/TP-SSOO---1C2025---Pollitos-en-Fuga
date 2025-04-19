package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/api"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

func main() {
	// 1. Cargar config y configurar logger
	global.InitGlobal() 
	defer global.LoggerKernel.CloseLogger()

	// 2. Crear y levantar server
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerKernel.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
	}()

	paqueteNuevo := paquetes.LeerConsola()	
	paquetes.GenerarYEnviarPaquete(paqueteNuevo, "127.0.0.1")
	
	//PASOS PROX: funcion crearPrimerProceso, funcion arrancar la planificacion corto y largo plazo, 
	select{}
} 