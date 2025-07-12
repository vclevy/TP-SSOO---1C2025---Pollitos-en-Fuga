package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	utilsIo "github.com/sisoputnfrba/tp-golang/io/utilsIo"
	"github.com/sisoputnfrba/tp-golang/io/global"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Paquete struct {
	PID    int `json:"pid"`
	TiempoDeBloqueo int `json:"tiempo"`
}

func RecibirPaquete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		//global.LoggerIo.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		//global.LoggerIo.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		//global.LoggerIo.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}

	global.LoggerIo.Log("IO recibió paquete: PID: "+strconv.Itoa(paquete.PID)+", Tiempo: "+strconv.Itoa(paquete.TiempoDeBloqueo), log.DEBUG)

	w.Write([]byte("IO recibió el paquete correctamente"))
}

func RecibirProceso(w http.ResponseWriter, r *http.Request) {
	var tarea estructuras.TareaDeIo
	if err := json.NewDecoder(r.Body).Decode(&tarea); err != nil {
		http.Error(w, "Error decodificando proceso", http.StatusBadRequest)
		return
	}

	global.LoggerIo.Log(fmt.Sprintf("IO: Recibí PID %d con tiempo %d", tarea.PID, tarea.TiempoEstimado), log.DEBUG)

	go utilsIo.IniciarIo(tarea)

	w.WriteHeader(http.StatusOK)
}
