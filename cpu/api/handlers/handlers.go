package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func Interrupcion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		PID int `json:"pid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	pidNuevo := data.PID // no se como haces palen para poner a ejecutar pero este es el nuevo pid
	pid := cpu.pidActual // ⚠️ Reemplazá esto por tu lógica real para obtener el PID
	pc := cpu.pcActual // ⚠️ Reemplazá esto por tu lógica real para obtener el PC

	global.Interrupcion = true
	global.LoggerCpu.Log(fmt.Sprintf("Interrupción recibida para PID %d (PC: %d)", pid, pc), log.DEBUG)

	// Respuesta JSON con pid y pc
	response := map[string]interface{}{
		"pid": pid,
		"pc":  pc,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

