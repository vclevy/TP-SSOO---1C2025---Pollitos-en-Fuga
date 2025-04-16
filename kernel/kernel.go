package main

import (
	"bytes"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/kernel/api"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	pack "github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

func main() {
	// 1. Cargar config y configurar logger
	global.InitGlobal() 
	defer global.Logger.CloseLogger()

	// 2. Crear y levantar server
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.Logger.Log("Error al iniciar el servidor: "+err_server.Error(), logger.ERROR)
		}
	}()

	// 3. Comunicarme con Memoria (u otros servicios)
	puertoMemoria := strconv.Itoa(global.KernelConfig.Port_Memory)
	url := "http://localhost:" + puertoMemoria + "/escribir"
	body := []byte("hola desde kernel")

	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Memoria:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)


	paquete := pack.LeerConsola()
	pack.GenerarYEnviarPaquete(paquete, "127.0.0.1", 8004 ) // Ip y Puerto de Cpu hardcodeado para probar

	// Bloqueo principal si es necesario (por ejemplo, esperar señales o input)
	select {}
}