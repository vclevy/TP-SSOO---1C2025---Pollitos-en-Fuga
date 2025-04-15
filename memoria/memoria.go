package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {
	// 1. Cargar config
	global.MemoriaConfig = config.CargarConfig[global.Config]("config/config.json")

	// 2. Inicializar logger
	global.Logger = log.ConfigurarLogger(global.MemoriaConfig.Log_file, global.MemoriaConfig.Log_level)
	defer global.Logger.CloseLogger()
	global.Logger.Log("Logger de memoria inicializado", log.DEBUG)

	// 3. Crear y levantar server
	s := api.CrearServer()
	fmt.Printf("🟢 Memoria prendida en http://localhost:%d\n", global.MemoriaConfig.Port_Memory)
	err := s.Iniciar()
	if err != nil {
		global.Logger.Log("Error al iniciar el servidor: "+err.Error(), log.ERROR)
	}
}
