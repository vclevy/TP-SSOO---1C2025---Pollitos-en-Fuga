package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
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
/* 
	pid := global.PCB_Actual.PID
	pc := global.PCB_Actual.PC */

	global.PCB_Interrupcion = data

	global.Interrupcion = true
/* 	global.LoggerCpu.Log(fmt.Sprintf("Interrupción recibida para PID %d (PC: %d)", global.PCB_Actual.PID, global.PCB_Actual.PC), log.DEBUG) */
	global.LoggerCpu.Log(("## Llega interrupción al puerto Interrupt"), log.DEBUG)

	/* response := estructuras.PCB{
		PID : pid,
		PC:  pc,
	} */

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(global.PCB_Actual.PID)
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

	global.PCB_Actual = data

	global.Interrupcion = true
	global.LoggerCpu.Log(fmt.Sprintf("Fue asignado un nuevo proceso con PID %d y PC: %d", global.PCB_Actual.PID, global.PCB_Actual.PC), log.DEBUG)

	response := estructuras.RespuestaCPU{
		PID : global.PCB_Actual.PID,
		PC:  global.PCB_Actual.PC,
		Motivo: global.Motivo,
		RafagaReal: global.Rafaga,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}