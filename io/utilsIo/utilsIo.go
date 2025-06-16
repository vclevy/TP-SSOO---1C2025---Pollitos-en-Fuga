package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	global.LoggerIo.Log(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %dms", solicitud.PID, solicitud.TiempoEstimado), log.INFO)

	// Simulación del proceso de E/S con sleep
	time.Sleep(time.Duration(solicitud.TiempoEstimado) * time.Millisecond)

	InformarFinalizacionDeIO(solicitud.PID)
}

func InformarFinalizacionDeIO(pid int) {
	mensaje := FinDeIO{
		Tipo: "FIN_IO",  
		PID:  pid,
	}

	body, err := json.Marshal(mensaje)
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error codificando mensaje de finalización IO: %v", err), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/finalizacionIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error creando request de finalización IO: %v", err), log.ERROR)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando mensaje de finalización IO: %v", err), log.ERROR)
		return
	}
	defer resp.Body.Close()

	global.LoggerIo.Log(fmt.Sprintf("PID %d - Finalización de IO notificada al Kernel", pid), log.INFO)
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
