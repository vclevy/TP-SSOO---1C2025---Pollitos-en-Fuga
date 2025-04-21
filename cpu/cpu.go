package main

import (
	"github.com/sisoputnfrba/tp-golang/cpu/api"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

func main() {
	// configurar logger e inicializar config
	global.InitGlobal() 
	defer global.LoggerCpu.CloseLogger()
	
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerCpu.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()

	for {
		paqueteNuevo := paquetes.LeerConsola()	
		paquetes.GenerarYEnviarPaquete(paqueteNuevo, "127.0.0.1")
	}
}