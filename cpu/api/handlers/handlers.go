package handlers

import (
	"io"
	"net/http"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"fmt"
	"bytes"
)

func HandshakeKernel(w http.ResponseWriter, r *http.Request) {
	datos := map[string]string{
		"id":     global.CpuID,
		"ip":     global.CpuConfig.Ip_Cpu,
		"puerto": fmt.Sprintf("%d", global.CpuConfig.Port_Cpu),
	}

	jsonData, _ := json.Marshal(datos)
	url := fmt.Sprintf("http://%s:%d/handshake", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando handshake al Kernel: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close()

	global.LoggerCpu.Log("✅ Handshake enviado al Kernel con éxito", log.INFO)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo respuesta del Kernel: " + err.Error(), log.ERROR)
		return
	}

	var datosRespuesta map[string]int
	err = json.Unmarshal(body, &datosRespuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando respuesta del Kernel: " + err.Error(), log.ERROR)
		return
	}

	pid := datosRespuesta["pid"]
	pc := datosRespuesta["pc"]

	global.LoggerCpu.Log(fmt.Sprintf(" Kernel respondió con PID: %d y PC: %d", pid, pc), log.INFO)
}

/* 
TODO:
? Ver lineas pid := datosRespuesta["pid"] y pc := datosRespuesta["pc"]
*/