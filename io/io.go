package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/io/global"
	"github.com/sisoputnfrba/tp-golang/io/utilsIo"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
)

func main() {

	global.InitGlobal()
	defer global.LoggerIo.CloseLogger()

	//paso 1: leer como parametro el nombre de la interaz io desde consola
	//para chequear que se pase el nombre
	if len(os.Args) < 2 {
		global.LoggerIo.Log("Tenés que pasar el nombre de una interfaz. Ejemplo: go run ./src/io.go teclado", log.ERROR) //Ver esto, porque no ejecuta así.. si no tiramos un cd directo
		return
	}

	nombreInterfaz := os.Args[1]

	fmt.Printf("Se conectó %s!\n", nombreInterfaz)

	infoIO := paquetes.Paquete{
		Mensajes:      []string{"Nombre de IO:" + nombreInterfaz, global.IoConfig.IPIo, strconv.Itoa(global.IoConfig.Port_Io)},
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

	// Simulación de la recepción de solicitudes de E/S
	for {
		// Aquí se puede simular una solicitud de I/O. En la práctica, esta parte debería estar en espera de peticiones del Kernel.
		solicitud := utilsIo.RespuestaKernel{
			PID:            respuesta.PID,
			TiempoEstimado: 1000, // Tiempo simulado de E/S
		}

		utilsIo.IniciarIo(solicitud)

		// Enviar una respuesta de finalización de I/O al Kernel
		// Este es el lugar donde puedes enviar una respuesta al Kernel indicando que la operación de I/O terminó.
		// Se puede enviar otro paquete con el status de finalización si es necesario.
	}
}
