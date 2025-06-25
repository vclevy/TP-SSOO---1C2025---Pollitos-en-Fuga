package main

import (
	"fmt"
	"os"
	"github.com/sisoputnfrba/tp-golang/cpu/api"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	utilsCpu "github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso: ./cpu <ID_CPU> <path_config>")
		return
	}

	idCPU := os.Args[1]
	configPath := os.Args[2]

	global.InitGlobal(idCPU, configPath)
	defer global.LoggerCpu.CloseLogger()

	if err := utilsCpu.HandshakeKernel(); err != nil {
		global.LoggerCpu.Log("Fallo el handshake con el Kernel: "+err.Error(), log.ERROR)
		os.Exit(1)
	}

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

