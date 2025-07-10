package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"time"
)

func Interrupcion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var data estructuras.PCB

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	global.PCB_Interrupcion = &data

	global.Interrupcion = true
	
	global.LoggerCpu.Log(("## Llega interrupción al puerto Interrupt"), log.INFO) //!! Interrupción Recibida - logObligatorio
}


func NuevoPCB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var data estructuras.PCB
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	global.PCB_Actual = &data
	global.LoggerCpu.Log(fmt.Sprintf("Fue asignado un nuevo proceso con PID %d y PC: %d", data.PID, data.PC), log.INFO)

	global.TiempoInicio = time.Now()
	w.WriteHeader(http.StatusOK)

	for utilsIo.CicloDeInstruccion() {
	}
}