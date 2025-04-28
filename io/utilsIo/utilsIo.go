package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO
type RespuestaKernel struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}

func HandshakeConKernel(paquete PaqueteHandshakeIO) (*RespuestaKernel, error) {

	global.LoggerIo.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), log.DEBUG)

	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerIo.Log("Error codificando paquete a JSON: "+err.Error(), log.ERROR)
		return nil, err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/handshakeIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel, err.Error()), log.ERROR)
		return nil, err 
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerIo.Log("Respuesta HTTP del Kernel: "+resp.Status, log.DEBUG)

	// Paso 6: Leer y procesar la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerIo.Log("Error leyendo la respuesta del Kernel: "+err.Error(), log.ERROR)
		return nil, err
	}

	// Paso 7: Deserializar la respuesta
	var respuesta RespuestaKernel //! ACA SALTA EL ERROR
	err = json.Unmarshal(respBody, &respuesta)
	if err != nil {
		global.LoggerIo.Log("Error parseando la respuesta del Kernel: "+err.Error(), log.ERROR)
		return nil, err
	}

	// Paso 8: Loguear la respuesta del Kernel en el log de IO
	global.LoggerIo.Log(fmt.Sprintf("Respuesta del Kernel: Status=%s | Detalle=%s | PID=%d | TiempoEstimado=%dms",
		respuesta.Status, respuesta.Detalle, respuesta.PID, respuesta.TiempoEstimado), log.DEBUG)

	return &respuesta, nil
}

func IniciarIo(solicitud RespuestaKernel) {
	// Log de inicio de E/S
	global.LoggerIo.Log(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %dms", solicitud.PID, solicitud.TiempoEstimado), log.INFO)

	// Simulación del proceso de E/S con sleep
	time.Sleep(time.Duration(solicitud.TiempoEstimado) * time.Millisecond)

	// Log de finalización de E/S
	global.LoggerIo.Log(fmt.Sprintf("## PID: %d - Fin de IO", solicitud.PID), log.INFO)
}