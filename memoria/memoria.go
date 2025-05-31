package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	//"github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

//CONEXIÓN;
//[KERNEL] ➜ Cliente (conectado a) [MEMORIA]
//[CPU]    ➜ Cliente (conectado a) [MEMORIA]

var memoriaUsuario []byte
var marcosLibres []bool

func main() {
	// configurar logger e inicializar config
	global.InitGlobal()
	defer global.LoggerMemoria.CloseLogger()

	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerMemoria.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()



	

	select{}
}

