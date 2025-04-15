package main

import (
	"bytes"
	"fmt"
	"net/http"
	utils "github.com/sisoputnfrba/tp-golang/utils/config" 
	"github.com/sisoputnfrba/tp-golang/io/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"bufio"
	"os"
	"time"
)

func main() {
	global.IoConfig = utils.CargarConfig[global.Config]("config/config.json")
	// 2. Inicializar logger
	global.Logger = logger.ConfigurarLogger(global.IoConfig.Log_File, global.IoConfig.LogLevel)
	defer global.Logger.CloseLogger()
	global.Logger.Log("Logger de io inicializado", logger.DEBUG)

	//paso 1: leer como parametro el nombre de la interaz io desde consola
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ingrese en nombre de la interfaz")
	nombreInterfaz, _ := reader.ReadString('\n') //falta que le pase esto al kernel
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)
	tiempoIO := time.Now().Format("2006-01-02 15:04:05")

	//paso 2: conectarse al kernel
	puertoKernel := strconv.Itoa(global.IoConfig.Port_Kernel)
	url := "http://localhost:"+ puertoKernel +"/escribir" 
	body := []byte("hola desde IO")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Kernel:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de kernel:", resp.Status)

	//loggear la conexio de io
	global.Logger.Log("##PID: "+ pidString + "- Inicio de IO - Tiempo " + tiempoIO, logger.DEBUG)


	

}
