package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	utilsIo "github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func HandshakeKernel(w http.ResponseWriter, r *http.Request) {
	datosEnvio := estructuras.HandshakeConCPU{
		ID   : global.CpuID,
		Puerto : global.CpuConfig.Port_Cpu,
		IP : global.CpuConfig.Ip_Cpu,
	}

	jsonData, _ := json.Marshal(datosEnvio)
	url := fmt.Sprintf("http://%s:%d/handshake", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //envia datosEnvio
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
	
	utilsIo.Fetch(pid,pc)
}
