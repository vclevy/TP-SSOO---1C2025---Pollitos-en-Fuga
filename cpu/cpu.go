package main

import (
	"os"
/* 	"github.com/sisoputnfrba/tp-golang/cpu/api" */
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	utilsCpu "github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
/* 	"github.com/sisoputnfrba/tp-golang/utils/logger" */
/* 	"github.com/sisoputnfrba/tp-golang/utils/paquetes" */
	"fmt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./cpu <ID_CPU>")
		return
	}

	idCPU := os.Args[1]
	global.InitGlobal(idCPU)

	utilsCpu.RealizarHandshakeConKernel()
	
	defer global.LoggerCpu.CloseLogger()
/* 	s := api.CrearServer() 
	
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerCpu.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
			os.Exit(1)
		}
	}()
*/
/* 	paqueteNuevo := paquetes.LeerConsola()	
	paquetes.GenerarYEnviarPaquete(paqueteNuevo, "127.0.0.1") */

	select{}
}