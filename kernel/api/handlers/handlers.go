package handlers

import (
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"encoding/json"
	"strings"
	"strconv"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int  `json:"codigo"`
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
		global.LoggerKernel.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerKernel.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerKernel.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}

	global.LoggerKernel.Log("Kernel recibió paquete desde IO - Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" | Código: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	// Simulación de asignación de PID y tiempo
	pid := 1234
	tiempoEstimado := 300

	respuesta := Respuesta{
		Status:        "OK",
		Detalle:       "Paquete procesado correctamente",
		PID:           pid,
		TiempoEstimado: tiempoEstimado,
	}

	global.LoggerKernel.Log("Kernel responde a IO: PID="+strconv.Itoa(pid)+", Tiempo="+strconv.Itoa(tiempoEstimado)+"ms", log.DEBUG)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}
