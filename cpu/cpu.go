package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"github.com/sisoputnfrba/tp-golang/cpu/api"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)


func main() {
	// configurar logger e inicializar config
	global.InitGlobal()
	
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerCpu.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()
		
		
	puertoMemoria := strconv.Itoa(global.CpuConfig.Port_Memory) //(string convert)
	//le habla a Memoria
	url := "http://localhost:"+ puertoMemoria + "/escribir" 
	body := []byte("hola desde CPU")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Memoria:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)

	puertoKernel := strconv.Itoa(global.CpuConfig.Port_Kernel)
	
	//le habla a Kernel
	urlKernel := "http://localhost:"+puertoKernel +"/escribir" 
	bodyParaKernel := []byte("hola desde CPU")
	respKernel, errKernel := http.Post(urlKernel, "text/plain", bytes.NewBuffer(bodyParaKernel))
	if errKernel != nil {
		fmt.Println("Error al mandar mensaje a kernel:", errKernel)
		return
	}
	defer respKernel.Body.Close()

	fmt.Println("Respuesta de kernel:", respKernel.Status)

	select {}

}