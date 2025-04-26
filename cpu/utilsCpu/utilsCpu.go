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

func RealizarHandshakeConKernel() {
	type datosEnvio struct {
		Id		string 	 `json:"id"`
		Ip  	string   `json:"ip"`
		Puerto	int		 `json:"puerto"`
	}

	type datosRespuesta struct {
		Pid		int		`json:"Pid"`
		Pc		int		`json:"Pc"`
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
}

func SolicitarInstruccionAMemoria(pid int, pc int) {
	type SolicitudInstruccion struct {
		Pid		int		`json:"Pid"`
		Pc		int		`json:"Pc"`
	}
	solicitudInstruccion := SolicitudInstruccion{
		Pid: pid,
		Pc:  pc,
	}

	jsonData, err := json.Marshal(solicitudInstruccion)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/solicitudInstruccion", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria con éxito", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var instruccionAEjecutar string
	err = json.Unmarshal(body, &instruccionAEjecutar)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return
	}

	global.LoggerCpu.Log(fmt.Sprintf("Memoria respondió con la instrucción: %s", instruccionAEjecutar), log.INFO)
}