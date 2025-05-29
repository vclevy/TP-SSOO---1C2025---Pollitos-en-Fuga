package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sisoputnfrba/tp-golang/io/global"
	utilsIO "github.com/sisoputnfrba/tp-golang/io/utilsIo"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
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

	infoIO := utilsIO.PaqueteHandshakeIO{
		NombreIO: nombreInterfaz,
		IPIO: global.IoConfig.IPIo,
		PuertoIO: global.IoConfig.Port_Io,
	}

	err := utilsIO.HandshakeConKernel(infoIO) //Usa el archivo utilsIo
	if err != nil {
		fmt.Println("Error al enviar paquete al Kernel:", err)
		return
	}

	//esta en los handlers -> cuando le llega una solicitud a la io, se iniciaIo con la funcion de utilsIO
	// Escuchar SIGINT o SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Recibida señal de cierre. Enviando desconexión a Kernel...")

		err := utilsIO.NotificarDesconexion(infoIO)
		if err != nil {
			fmt.Println("Error notificando desconexión:", err)
		} else {
			fmt.Println("Desconexión notificada con éxito.")
		}
		os.Exit(0)
	}()

	select {}
}