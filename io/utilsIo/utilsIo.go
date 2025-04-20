package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/io/global"
)


type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
}
type RespuestaKernel struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}


func EnviarPaqueteAKernel(paquete Paquete, ip string) (*RespuestaKernel, error) {
	// Paso 1: Validar que haya mensajes en el paquete
	if len(paquete.Mensajes) == 0 {
		global.LoggerIo.Log("No se ingresaron mensajes para enviar.", "ERROR")
		return nil, fmt.Errorf("no se ingresaron mensajes para enviar")
	}

	// Paso 2: Log de paquete a enviar
	global.LoggerIo.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), "DEBUG")

	// Paso 3: Convertir el paquete a JSON
	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerIo.Log("Error codificando paquete a JSON: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/responder", ip, paquete.PuertoDestino)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", ip, paquete.PuertoDestino, err.Error()), "ERROR")
		return nil, err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerIo.Log("Respuesta HTTP del Kernel: "+resp.Status, "DEBUG")

	// Paso 6: Leer y procesar la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerIo.Log("Error leyendo la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 7: Deserializar la respuesta
	var respuesta RespuestaKernel
	err = json.Unmarshal(respBody, &respuesta)
	if err != nil {
		global.LoggerIo.Log("Error parseando la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 8: Loguear la respuesta del Kernel en el log de IO
	global.LoggerIo.Log(fmt.Sprintf("Respuesta del Kernel: Status=%s | Detalle=%s | PID=%d | TiempoEstimado=%dms",
		respuesta.Status, respuesta.Detalle, respuesta.PID, respuesta.TiempoEstimado), "DEBUG")

	// Devolver la respuesta
	return &respuesta, nil
}