package main

import (
	"fmt"
	"os"
	"github.com/sisoputnfrba/tp-golang/cpu/api"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./cpu <ID_CPU>")
		return
	}

	idCPU := os.Args[1]
	global.InitGlobal(idCPU)

	defer global.LoggerCpu.CloseLogger()
	s := api.CrearServer()

	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerCpu.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
			os.Exit(1)
		}
	}()
	
	select {}	
}
