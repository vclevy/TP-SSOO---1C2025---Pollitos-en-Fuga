package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)
/* type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
}
*/
func RealizarHandshakeConKernel() {
	/* 	datosEnvio := map[string]string{
		"id":     global.CpuID,
		"ip":     global.CpuConfig.IPCpu,
		"puerto": fmt.Sprintf("%d", global.CpuConfig.Port_CPU),
	} */

	type datosEnvio struct {
		Id		string 	 `json:"id"`
		Ip  	string   `json:"ip"`
		Puerto	int		 `json:"puerto"`
	}

	type datosRespuesta struct {
		Pid		int		`json:"Pid_kernel"`
		Pc		int		`json:"Pc_kernel"`
	}
 	
	//envio
	var envio datosEnvio

	jsonData, err := json.Marshal(envio)
	if err != nil {
		global.LoggerCpu.Log("Error serializando handshake: "+err.Error(), log.ERROR)
		return
	}
	
	url := fmt.Sprintf("http://%s:%d/handshake", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando handshake al Kernel: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Handshake enviado al Kernel con éxito", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var respuesta datosRespuesta
	err = json.Unmarshal(body, &respuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando respuesta del Kernel: "+err.Error(), log.ERROR)
		return
	}

	global.LoggerCpu.Log(fmt.Sprintf("Kernel respondió con PID: %d y PC: %d", respuesta.Pid, respuesta.Pc), log.INFO)






	
/* 
	if err != nil {
		global.LoggerCpu.Log("Error leyendo respuesta del Kernel: " + err.Error(), log.ERROR)
		return
	} */
/* 
	var datosRespuesta map[string]int */
	/* err = json.Unmarshal(body, &datosRespuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando respuesta del Kernel: " + err.Error(), log.ERROR)
		return
	} */
/* 
	pid := datosRespuesta["pid"]
	pc := datosRespuesta["pc"]
 */
	/* global.LoggerCpu.Log(fmt.Sprintf(" Kernel respondió con PID: %d y PC: %d", pid, pc), log.INFO) */
}


/* type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
}
type RespuestaKernel struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}

type RespuestaMemoria struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
} */

/* func EnviarPaqueteAKernel(paquete Paquete, ip string) (*RespuestaKernel, error) {
	// Paso 1: Validar que haya mensajes en el paquete
	if len(paquete.Mensajes) == 0 {
		global.LoggerCpu.Log("No se ingresaron mensajes para enviar.", "ERROR")
		return nil, fmt.Errorf("no se ingresaron mensajes para enviar")
	}

	// Paso 2: Log de paquete a enviar
	global.LoggerCpu.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), "DEBUG")

	// Paso 3: Convertir el paquete a JSON
	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerCpu.Log("Error codificando paquete a JSON: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/responder", ip, 8001)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerCpu.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", ip, 8001, err.Error()), "ERROR")
		return nil, err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerCpu.Log("Respuesta HTTP del Kernel: "+resp.Status, "DEBUG")

	// Paso 6: Leer y procesar la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 7: Deserializar la respuesta
	var respuesta RespuestaKernel
	err = json.Unmarshal(respBody, &respuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 8: Loguear la respuesta del Kernel en el log de CPU
	global.LoggerCpu.Log(fmt.Sprintf("Respuesta del Kernel: Status=%s | Detalle=%s | PID=%d | TiempoEstimado=%dms",
		respuesta.Status, respuesta.Detalle, respuesta.PID, respuesta.TiempoEstimado), "DEBUG")

	// Devolver la respuesta
	return &respuesta, nil
}
 */
/* func EnviarPaqueteAMemoria(paquete Paquete, ip string) (*RespuestaMemoria, error) {
	// Paso 1: Validar que haya mensajes en el paquete
	if len(paquete.Mensajes) == 0 {
		global.LoggerCpu.Log("No se ingresaron mensajes para enviar.", "ERROR")
		return nil, fmt.Errorf("no se ingresaron mensajes para enviar")
	}

	// Paso 2: Log de paquete a enviar
	global.LoggerCpu.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), "DEBUG")

	// Paso 3: Convertir el paquete a JSON
	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerCpu.Log("Error codificando paquete a JSON: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 4: Enviar el paquete a Memoria (POST)
	url := fmt.Sprintf("http://%s:%d/responder", ip, 8002)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerCpu.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", ip, 8002, err.Error()), "ERROR")
		return nil, err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerCpu.Log("Respuesta HTTP del Memoria: "+resp.Status, "DEBUG")

	// Paso 6: Leer y procesar la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo la respuesta del Memoria: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 7: Deserializar la respuesta
	var respuesta RespuestaMemoria
	err = json.Unmarshal(respBody, &respuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando la respuesta del Memoria: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 8: Loguear la respuesta del Memoria en el log de CPU
	global.LoggerCpu.Log(fmt.Sprintf("Respuesta de Memoria: Status=%s | Detalle=%s | PID=%d | TiempoEstimado=%dms",
		respuesta.Status, respuesta.Detalle, respuesta.PID, respuesta.TiempoEstimado), "DEBUG")

	// Devolver la respuesta
	return &respuesta, nil
} */