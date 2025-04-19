package utilsKernel

import (
	"fmt"
	"net/http"
	"sort"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PCB struct {
	PID int
	PC  int
	ME  map[string]int // Métricas de Estado (ej: "Ready": 3, "Running": 5)
	MT  map[string]int // Métricas de Tiempo por Estado (ej: "Ready": 120, "Running": 300)
}

func NuevoPCB(pid int) *PCB {
	return &PCB{
		PID: pid,
		PC:  0,
		ME:  make(map[string]int),
		MT:  make(map[string]int),
	}
}

type Proceso struct {
	PCB
	MemoriaRequerida int
}

func CrearProceso(pid int, pseudoCodigo string, tamanio int) {
	pcb := NuevoPCB(pid)
	//TODO Parte de pedir memoria
	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
	}

//	log := fmt.Sprintf("## (%d:0) Se crea el Proceso - Estado: NEW", pcb.PID)
//	global.Logger.Log(log, log.INFO)

	switch global.AlgoritmoLargoPlazo {
	case "FIFO":
		if len(global.ColaNew) == 0 {
			// Intentamos iniciarlo directo
			if SolicitarMemoria(tamanio) == http.StatusOK {
				PasarPseudocodigoAMemoria(proceso)
				//si funciona, pasa a ready
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", pcb.PID), log.INFO)
				return
			}
		}
		// Si no, lo ponemos en la cola NEW
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (FIFO)", pcb.PID), log.INFO) //No es nuestro log obligatorio

	case "CHICO":
		if SolicitarMemoria(tamanio) == http.StatusOK {
			PasarPseudocodigoAMemoria(proceso)
			global.ColaReady = append(global.ColaReady, proceso)
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", pcb.PID), log.INFO)
			return
		}
		// No hubo espacio → encolamos
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (CHICO)", pcb.PID), log.INFO)
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
			PasarPseudocodigoAMemoria(proceso)
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaNew = global.ColaNew[1:] // sacamos el primero
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (FIFO)", proceso.PID), log.INFO)
		}

	case "CHICO":
		// ordenamos por tamaño
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				PasarPseudocodigoAMemoria(proceso)
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (CHICO)", proceso.PID), log.INFO)
			} else {
				nuevaCola = append(nuevaCola, proceso)
			}
		}
		global.ColaNew = nuevaCola
	}
}
