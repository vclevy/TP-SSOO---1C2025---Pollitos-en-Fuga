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
	global.LoggerKernel.Log("Kernel recibió paquete: Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" Codigo: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	w.Write([]byte("Kernel recibió el paquete correctamente"))
}