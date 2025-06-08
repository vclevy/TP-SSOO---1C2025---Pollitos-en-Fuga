package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	pid := global.PCB_Actual.PID
	pc := global.PCB_Actual.PC

	global.PCB_Actual = data

	global.Interrupcion = true
	global.LoggerCpu.Log(fmt.Sprintf("Interrupción recibida para PID %d (PC: %d)", pid, pc), log.DEBUG)

	response := estructuras.PCB{
		PID : pid,
		PC:  pc,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func NuevoPCB(w http.ResponseWriter, r *http.Request) {
    var pcb estructuras.PCB

    err := json.NewDecoder(r.Body).Decode(&pcb)
    if err != nil {
        http.Error(w, "Error en el body: "+err.Error(), http.StatusBadRequest)
        return
    }

    global.PCB_Actual = pcb

    tiempoInicio := time.Now()

    // Ejecutar ciclo de instrucciones, puede estar en otra función (por ejemplo: cpu.CicloDeInstruccion)
    Motivo := EjecutarProceso()  // implementá tu ciclo de instrucción aquí, que retorna el motivo de finalización

    tiempoRafaga := time.Since(tiempoInicio).Seconds()

    respuesta := estructuras.RespuestaCPU{
        PID: pcb.PID,
        PC: global.PCB_Actual.PC,
        Motivo: Motivo,
        RafagaReal: tiempoRafaga,
    }

    jsonData, err := json.Marshal(respuesta)
    if err != nil {
        http.Error(w, "Error serializando respuesta: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Enviar respuesta al Kernel (post a /respuestaCPU)
    urlKernel := fmt.Sprintf("http://%s:%d/respuestaCPU", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)
    resp, err := http.Post(urlKernel, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        http.Error(w, "Error enviando respuesta al kernel: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    w.WriteHeader(http.StatusOK)
}

