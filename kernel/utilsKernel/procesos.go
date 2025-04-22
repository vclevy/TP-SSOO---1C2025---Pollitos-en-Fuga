package utilsKernel

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PCB = global.PCB
type Proceso = global.Proceso

type Estado string

func CrearProceso(tamanio int) Proceso {
	pcb := global.NuevoPCB()
	ActualizarEstadoPCB(pcb, "New")

	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
	}

	global.LoggerKernel.Log(fmt.Sprintf("## (%d:0) Se crea el proceso - Estado: NEW", pcb.PID), log.INFO)
	return proceso
}

func PlanificarProcesoLargoPlazo(proceso Proceso) {
	switch global.AlgoritmoLargoPlazo {
	case "FIFO":
		if len(global.ColaNew) == 0 {
			if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				//TODO PasarPseudocodigoAMemoria(proceso)
				ActualizarEstadoPCB(&proceso.PCB, "Ready")
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", proceso.PCB.PID), log.INFO)
				return
			}
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (FIFO)", proceso.PCB.PID), log.INFO)

	case "CHICO":
		if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
			//TODO PasarPseudocodigoAMemoria(proceso)
			ActualizarEstadoPCB(&proceso.PCB, "Ready")
			global.ColaReady = append(global.ColaReady, proceso)
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", proceso.PCB.PID), log.INFO)
			return
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (CHICO)", proceso.PCB.PID), log.INFO)
	}
}


func IntentarInicializarDesdeNew() {
	if len(global.ColaNew) == 0 {
		return
	}

	switch global.AlgoritmoLargoPlazo {
	case "FIFO":
		proceso := global.ColaNew[0]
		if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
			//TODO PasarPseudocodigoAMemoria(proceso)
			ActualizarEstadoPCB(&proceso.PCB, "Ready")
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaNew = global.ColaNew[1:]
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (FIFO)", proceso.PID), log.INFO)
		}

	case "CHICO":
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				//TODO PasarPseudocodigoAMemoria(proceso)
				ActualizarEstadoPCB(&proceso.PCB, "Ready")
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (CHICO)", proceso.PID), log.INFO)
			} else {
				nuevaCola = append(nuevaCola, proceso)
			}
		}
		global.ColaNew = nuevaCola
	}
}

func SolicitarMemoria(tamanio int) int {
	cliente := &http.Client{}
	endpoint := "tamanioProceso/" + strconv.Itoa(tamanio)
	url := fmt.Sprintf("http://%s:%d/%s", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusInternalServerError
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer respuesta.Body.Close()

	return respuesta.StatusCode
}

func ActualizarEstadoPCB(pcb *PCB, nuevoEstado string) {
	ahora := time.Now()

	// Si ya tenía un estado previo, calculamos tiempo en ese estado
	if pcb.UltimoEstado != "" {
		duracion := int(ahora.Sub(pcb.InicioEstado).Milliseconds())
		pcb.MT[pcb.UltimoEstado] += duracion
	}

	// Aumenta contador de veces en el nuevo estado
	pcb.ME[nuevoEstado] += 1

	// Actualiza último estado y momento de entrada
	pcb.UltimoEstado = nuevoEstado
	pcb.InicioEstado = ahora
}
