package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sisoputnfrba/tp-golang/io/api"
	"github.com/sisoputnfrba/tp-golang/io/global"
	utilsIO "github.com/sisoputnfrba/tp-golang/io/utilsIo"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Uso: go run ./src/io.go <dispositivo> <path_config>")
		return
	}

	configPath := os.Args[2]
	global.InitGlobal(configPath)
	defer global.LoggerIo.CloseLogger()
	
s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerIo.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
	}()
	nombreInterfaz := os.Args[1]

	fmt.Printf("Se conectó %s!\n", nombreInterfaz)

	infoIO := utilsIO.PaqueteHandshakeIO{
		NombreIO: strings.ToLower(nombreInterfaz),
		IPIO: global.IoConfig.IPIo,
		PuertoIO: global.IoConfig.Port_Io,
	}

	err := utilsIO.HandshakeConKernel(infoIO) 
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