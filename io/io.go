package main

import (
	"fmt"
	"os"
	"strconv"
	"github.com/sisoputnfrba/tp-golang/io/utilsIo"
	"github.com/sisoputnfrba/tp-golang/io/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
	//"github.com/sisoputnfrba/tp-golang/io/api"
)

func main() {

	global.InitGlobal() 
	defer global.LoggerIo.CloseLogger()

	//paso 1: leer como parametro el nombre de la interaz io desde consola
	//para chequear que se pase el nombre
	if len(os.Args) < 2 {
		global.LoggerIo.Log("Tenés que pasar el nombre de una interfaz. Ejemplo: go run ./src/io.go teclado", logger.ERROR) //Ver esto, porque no ejecuta así.. si no tiramos un cd directo
        return
    }

    nombreInterfaz := os.Args[1]

    fmt.Printf("Se conectó %s!\n", nombreInterfaz)


	infoIO := paquetes.Paquete{
		Mensajes:      []string{"Nombre de IO:" + nombreInterfaz, global.IoConfig.IPIo ,strconv.Itoa(global.IoConfig.Port_Io)},
		Codigo:        200, //??????
		PuertoDestino: global.IoConfig.Port_Kernel,
	}
	
	respuesta, err := utilsIo.EnviarPaqueteAKernel(utilsIo.Paquete(infoIO), global.IoConfig.IPKernel) //Usa el archivo utilsIo
	if err != nil {
	fmt.Println("Error al enviar paquete al Kernel:", err)
	return
}

fmt.Println("Respuesta del Kernel:")
fmt.Println("Status:", respuesta.Status)
fmt.Println("Detalle:", respuesta.Detalle)

	
	// s := api.CrearServer()
	// go func() {
	// 	err_server := s.Iniciar()
	// 	if err_server != nil {
	// 		global.LoggerIo .Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
	// 	}
	// }()
	
	//el kernel le va a mandar un paquete con el TIEMPO estimado que va a tardar y el PID asignado
	//api.RecibirPaquete() ..
	//global.LoggerIo.Log("##PID: "+ pidString + "- Inicio de IO - Tiempo " + tiempoIO, log.DEBUG) 
	// //Al momento de recibir una petición del Kernel, el módulo deberá iniciar un usleep por el tiempo indicado en la request
	// time.Sleep(time.Duration(tiempoIO) * time.Millisecond)
	// global.LoggerIo.Log("##PID: "+ pidString + "- FIn de IO", log.DEBUG)


}


