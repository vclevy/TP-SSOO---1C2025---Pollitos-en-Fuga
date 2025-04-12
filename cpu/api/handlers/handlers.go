package handlers

import (
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Paquete struct {
	Mensaje string `json:"mensaje"`
	Codigo  int    `json:"codigo"`
}


func RecibirPaqueteDeKernel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		global.LoggerCpu.Log("Se intentó acceder con un método no permitido", logger.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerCpu.Log("Error leyendo el cuerpo del request: "+err.Error(), logger.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerCpu.Log("Error al parsear el paquete JSON: "+err.Error(), logger.DEBUG)
		return
	}

	global.LoggerCpu.Log("CPU recibió paquete de Kernel: "+paquete.Mensaje, logger.DEBUG)

	w.Write([]byte("CPU recibió el paquete correctamente"))
}


