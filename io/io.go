package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/global"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

func main() {
	global.IoConfig = config.CargarConfig[global.Config]("config/config.json")
	// 2. Inicializar logger
	global.Logger = log.ConfigurarLogger(global.IoConfig.Log_File, global.IoConfig.LogLevel)
	defer global.Logger.CloseLogger()
	global.Logger.Log("Logger de io inicializado", log.DEBUG)

	//paso 1: leer como parametro el nombre de la interaz io desde consola
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Ingrese en nombre de la interfaz")
	nombreInterfaz, _ := reader.ReadString('\n') //falta que le pase esto al kernel
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)
	tiempoIO := time.Now().Format("2006-01-02 15:04:05")

	infoIO := paquetes.Paquete{
		Mensajes:      []string{"Nombre de IO: " + nombreInterfaz, strconv.Itoa(global.IoConfig.Port_Io)}, //falta pasarla la ip, pero no se de quien
		Codigo:        200,//no se que va aca
		PuertoDestino: global.IoConfig.Port_Kernel, //puerto kernel
	}
	paquetes.GenerarYEnviarPaquete(infoIO,global.IoConfig.IPKernel,global.IoConfig.Port_Kernel)
	global.Logger.Log("##PID: "+ pidString + "- Inicio de IO - Tiempo " + tiempoIO, log.DEBUG)
	global.Logger.Log("##PID: "+ pidString + "- FIn de IO", log.DEBUG)



	


	

}
