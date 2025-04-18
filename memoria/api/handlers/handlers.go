package handlers

import (
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
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
		global.LoggerMemoria.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerMemoria.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerMemoria.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}
	global.LoggerMemoria.Log("Memoria recibió paquete: Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" Codigo: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	w.Write([]byte("Memoria recibió el paquete correctamente"))
}