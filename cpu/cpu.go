package main

import (
	"fmt"
	"os"
	"github.com/sisoputnfrba/tp-golang/cpu/api"
	"github.com/sisoputnfrba/tp-golang/cpu/api/handlers"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	utilsIo "github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var pidActual int
var pcActual int

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

	PCB, err := utilsIo.HandshakeKernel()
	if err != nil {
		fmt.Println("Error en handshake con el Kernel:", err)
		return
	}

	pidActual = PCB.Pid
	pcActual = PCB.Pc
	
	utilsIo.ConfigMMU()
	utilsIo.CicloDeInstruccion(pidActual,pcActual)
	
	select {}

	
}
