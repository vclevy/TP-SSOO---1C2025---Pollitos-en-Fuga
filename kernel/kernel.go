package main

import (
	"bytes"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/kernel/api"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

func main() {
	
	// 1. Cargar config
	global.KernelConfig =  utils.CargarConfig[global.Config]("config/config.json")
	
	puertoMemoria := strconv.Itoa(global.KernelConfig.Port_Memory) //(string convert)
	url := "http://localhost:"+ puertoMemoria+"/escribir" 
	body := []byte("hola desde kernel")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
	fmt.Println("Error al mandar mensaje a Memoria:", err)
	return
	 }
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)
	// 2. Inicializar logger
	global.Logger = logger.ConfigurarLogger(global.KernelConfig.Log_file, global.KernelConfig.LogLevel)
	defer global.Logger.CloseLogger()
	global.Logger.Log("Logger de memoria inicializado", logger.DEBUG)

	// 3. Crear y levantar server
	s := api.CrearServer()
	fmt.Printf("🟢 Kernel prendido en http://localhost:%d\n", global.KernelConfig.Port_Kernel)
	err_server := s.Iniciar()
	if err_server != nil {
		global.Logger.Log("Error al iniciar el servidor: "+err_server.Error(), logger.ERROR)
	}
	


}
