package handlers

import (
	"io"
	"net/http"
	"encoding/json"
	"strings"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
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