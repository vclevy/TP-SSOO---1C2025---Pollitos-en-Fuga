package main

import (
	"fmt"
	"os"
	"github.com/sisoputnfrba/tp-golang/memoria/api"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

//CONEXIÓN;
//[KERNEL] ➜ Cliente (conectado a) [MEMORIA]
//[CPU]    ➜ Cliente (conectado a) [MEMORIA]

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Uso: ./memoria <path_config>")
		return
	}

	configPath := os.Args[1]
	global.InitGlobal(configPath)
	defer global.LoggerMemoria.CloseLogger()

	utilsMemoria.InicializarMemoria()

	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerMemoria.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()

	

	select{}
}

