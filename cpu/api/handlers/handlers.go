package handlers

import (
	"io"
	"net/http"
	"encoding/json"
	"strings"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"fmt"
	"bytes"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
}

type Respuesta struct {
	Status        string `json:"status"`
	Detalle       string `json:"detalle"`
	PID           int    `json:"pid"`
	TiempoEstimado int   `json:"tiempo_estimado"`
}

func RecibirPaquete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		global.LoggerCpu.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerCpu.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerCpu.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}
	global.LoggerCpu.Log("CPU recibió paquete: Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" Codigo: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	w.Write([]byte("CPU recibió el paquete correctamente"))
}

func RealizarHandshakeConKernel() {
	datos := map[string]string{
		"id":     global.CpuID,
		"ip":     global.CpuConfig.IPCpu,
		"puerto": fmt.Sprintf("%d", global.CpuConfig.Port_CPU),
	}

	jsonData, _ := json.Marshal(datos)
	url := fmt.Sprintf("http://%s:%d/handshake", global.CpuConfig.IPKernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando handshake al Kernel: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close()

	global.LoggerCpu.Log("✅ Handshake enviado al Kernel con éxito", log.INFO)
}