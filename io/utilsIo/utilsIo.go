package utilsIo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/global"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO
type FinDeIO = estructuras.FinDeIO

func HandshakeConKernel(paquete PaqueteHandshakeIO) error {
	global.LoggerIo.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), log.DEBUG)

	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerIo.Log("Error codificando paquete a JSON: "+err.Error(), log.ERROR)
		return err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/handshakeIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel, err.Error()), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerIo.Log("Respuesta HTTP del Kernel: "+resp.Status, log.DEBUG)

	return nil
}

func IniciarIo(solicitud estructuras.TareaDeIo) {
	// Log de inicio de E/S
	global.LoggerIo.Log(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %dms", solicitud.PID, solicitud.TiempoEstimado), log.INFO)

	// Simulación del proceso de E/S con sleep
	time.Sleep(time.Duration(solicitud.TiempoEstimado) * time.Millisecond)

	InformarFinalizacionDeIO(solicitud.PID)
}

func InformarFinalizacionDeIO(pid int){
	
	tipo, err := strconv.Atoi(leerOpcionConsola())
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error convirtiendo opción a entero: %v", err), log.ERROR)
		return
	}

	var mensaje FinDeIO

	if tipo==1{
		mensaje.Tipo = "FIN_IO"
	} else if tipo==2{
		mensaje.Tipo = "DESCONEXION_IO"
	}

	body, err := json.Marshal(mensaje)
    if err != nil {
        global.LoggerIo.Log(fmt.Sprintf("Error codificando mensaje: %v", err), log.ERROR)
        return
    }

    url := fmt.Sprintf("http://%s:%d/finalizacionIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        global.LoggerIo.Log(fmt.Sprintf("Error creando request: %v", err), log.ERROR)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        global.LoggerIo.Log(fmt.Sprintf("Error enviando mensaje: %v", err), log.ERROR)
        return
    }
    defer resp.Body.Close()

    global.LoggerIo.Log(fmt.Sprintf("## PID: <%d> - Fin de IO notificado", pid), log.INFO)
}


func leerOpcionConsola() string {
	reader := bufio.NewReader(os.Stdin)

	for { // Bucle infinito hasta que se ingrese 1 o 2
		fmt.Println("\nSeleccione el tipo de mensaje:")
		fmt.Println("1 - FIN_IO")
		fmt.Println("2 - DESCONEXION_IO")
		fmt.Print("Ingrese opción (1/2): ")

		opcion, _ := reader.ReadString('\n')
		opcion = strings.TrimSpace(opcion) // Elimina espacios y saltos de línea

		if opcion == "1" || opcion == "2" {
			return opcion // Retorna la opción válida
		}

		// Mensaje de error y reintento
		fmt.Printf("\n** ERROR: '%s' no es válido. Solo se permite 1 o 2 **\n", opcion)
	}
}


func NotificarDesconexion(info PaqueteHandshakeIO) error {
	url := fmt.Sprintf("http://%s:%d/finalizacionIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)

	req, err := http.NewRequest("POST", url, nil) // sin body = desconexión
	if err != nil {
		return err
	}

	req.Header.Set("X-IO-Nombre", info.NombreIO)
	req.Header.Set("X-IO-Puerto", strconv.Itoa(info.PuertoIO))
	req.Header.Set("X-IO-IP", info.IPIO)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("falló desconexión: %s", resp.Status)
	}

	return nil
}
