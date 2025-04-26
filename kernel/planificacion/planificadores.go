package planificacion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	//estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
)

const (
	NEW          string = "NEW"
	READY        string = "READY"
	EXEC          string = "EXEC"
	EXIT         string = "EXIT"
	BLOCKED         string = "BLOCKED"
	SUSP_READY   string = "SUSP READY"
	SUSP_BLOCKED string = "SUSP BLOCKED"
)

var estado = []string{
	NEW,
	READY,
	EXEC,
	EXIT,
	BLOCKED,
	SUSP_READY,
	SUSP_BLOCKED,
}

type PCB = global.PCB
type Proceso = global.Proceso

func CrearProceso(tamanio int, archivoPseudoCodigo string) Proceso {
	pcb := global.NuevoPCB()
	ActualizarEstadoPCB(pcb, NEW)

	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
		ArchivoPseudo:    archivoPseudoCodigo,
	}

	global.LoggerKernel.Log(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID), log.INFO) //! LOG OBLIGATORIO: Creacion de Proceso
	global.ColaNew = append(global.ColaNew, proceso)
	return proceso
}

func IniciarPlanificadorLargoPlazo() {
	go func() {
		<-global.InicioPlanificacionLargoPlazo
		global.LoggerKernel.Log("Iniciando planificación de largo plazo...", log.INFO)

		for {
			// 🧹 Finalización de procesos
			if len(global.ColaExit) > 0 {

				p := global.ColaExit[0]
				FinalizarProceso(&p)
				global.ColaExit = global.ColaExit[1:]

				// Intentar inicializar proceso desde NEW
				if IntentarInicializarDesdeNew() {
					continue
				}
			}

			// 🕒 Si no hay nada que hacer
			if len(global.ColaNew) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Planificación de procesos en NEW
			switch global.ConfigKernel.SchedulerAlgorithm {
			case "FIFO":
				proceso := global.ColaNew[0]
				if SolicitarMemoria(proceso.MemoriaRequerida) {
					global.ColaNew = global.ColaNew[1:] // Quitar el primer proceso de NEW
					ActualizarEstadoPCB(&proceso.PCB, READY)
					global.ColaReady = append(global.ColaReady, proceso)
					global.LoggerKernel.Log(fmt.Sprintf("## (%d) Pasa del estado NEW al estado READY", proceso.PCB.PID), log.INFO)

					EvaluarDesalojo(proceso)
				}
			case "CHICO":
				// Ordenar por menor memoria requerida
				ordenada := make([]Proceso, len(global.ColaNew))
				copy(ordenada, global.ColaNew)
				sort.Slice(ordenada, func(i, j int) bool {
					return ordenada[i].MemoriaRequerida < ordenada[j].MemoriaRequerida
				})

				// Buscar el primer proceso con memoria disponible
				for _, proc := range ordenada {
					if SolicitarMemoria(proc.MemoriaRequerida) {
						// Eliminarlo de la cola original
						for i, p := range global.ColaNew {
							if p.PCB.PID == proc.PCB.PID {
								global.ColaNew = append(global.ColaNew[:i], global.ColaNew[i+1:]...)
								break
							}
						}
						ActualizarEstadoPCB(&proc.PCB, READY)
						global.ColaReady = append(global.ColaReady, proc)
						global.LoggerKernel.Log(fmt.Sprintf("## (%d) Pasa del estado NEW al estado READY", proc.PCB.PID), log.INFO)

						EvaluarDesalojo(proc)

						break
					}
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IntentarInicializarDesdeNew() bool {
	if len(global.ColaNew) == 0 {
		return false
	}

	moved := false

	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso := global.ColaNew[0]
		if SolicitarMemoria(proceso.MemoriaRequerida) {
			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaNew = global.ColaNew[1:]
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (FIFO)", proceso.PCB.PID), log.INFO)
			moved = true
		}

	case "CHICO":
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) {
				ActualizarEstadoPCB(&proceso.PCB, READY)
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (CHICO)", proceso.PCB.PID), log.INFO)
				moved = true
			} else {
				nuevaCola = append(nuevaCola, proceso)
			}
		}
		global.ColaNew = nuevaCola
	}

	return moved
}

func SolicitarMemoria(tamanio int) bool {
	cliente := &http.Client{}
	endpoint := "verificarEspacioDisponible/" + strconv.Itoa(tamanio) // Correcto si 'tamanioProceso' es el handler de Memoria
	url := fmt.Sprintf("http://%s:%d/%s", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory, endpoint)

	// Crear la solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false // Error al crear la solicitud
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return false // Error al enviar la solicitud
	}
	defer respuesta.Body.Close()

	// Si la respuesta es 200 OK, entonces retornamos true (éxito)
	if respuesta.StatusCode == http.StatusOK {
		return true
	}

	return false
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

func InformarFinAMemoria(pid int) error {
	url := "http://" + global.ConfigKernel.IPMemory + ":" + strconv.Itoa(global.ConfigKernel.Port_Memory) + "/finalizar-proceso"

	data := map[string]int{"pid": pid}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("memoria devolvió error")
	}
	return nil
}

func FinalizarProceso(p *Proceso) {

	ActualizarEstadoPCB(&p.PCB, EXIT)

	err := InformarFinAMemoria(p.PID)
	if err != nil {
		// Manejo de error
		return
	}

	LoguearMetricas(p) // (log obligatorio)
	global.ColaExecuting = filtrarCola(global.ColaExecuting, p)
	global.ColaExit = append(global.ColaExit, *p)

}

func filtrarCola(cola []global.Proceso, target *Proceso) []global.Proceso {
	nueva := []global.Proceso{}
	for _, proc := range cola {
		if proc.PID != target.PID {
			nueva = append(nueva, proc)
		}
	}
	return nueva
}

func LoguearMetricas(p *Proceso) {
	global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Finaliza el proceso", p.PID), log.INFO) //! LOG OBLIGATORIO: Fin de Proceso

	msg := fmt.Sprintf("## (%d) - Métricas de estado:", p.PID)

	for _, unEstado := range estado {
		count := p.ME[unEstado]
		tiempo := p.MT[unEstado]
		msg += fmt.Sprintf(" %s (%d) (%d),", unEstado, count, tiempo)
	}

	// Eliminar la coma final
	msg = msg[:len(msg)-1]

	global.LoggerKernel.Log(msg, log.INFO) //! LOG OBLIGATORIO: Metricas de Estado
}

/*	func EnviarProcessDataAMemoria(proceso Proceso, archPseudo string){
	pid := proceso.PCB.PID
	pseudoCodigo := proceso
	return
} */

func IniciarPlanificadorCortoPlazo() {
	go func() {
		for {
			SeleccionarYDespacharProceso()
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func SeleccionarYDespacharProceso() {
	// No hacer nada si no hay procesos READY o no hay CPUs libres
	if len(global.ColaReady) == 0 || !HayCPUDisponible() {
		return
	}

	var proceso Proceso

	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso = global.ColaReady[0]
		global.ColaReady = global.ColaReady[1:]

	case "SJF":
		proceso = seleccionarProcesoSJF(false)

	case "SRTF":
		proceso = seleccionarProcesoSJF(true)
	}

	EnviarADispatch(proceso.PCB.PID)
	ActualizarEstadoPCB(&proceso.PCB, EXEC)
	global.ProcesoEnEjecucion = proceso
}

func seleccionarProcesoSJF(desalojo bool) Proceso {
	// Ordenar por estimación de ráfaga
	sort.Slice(global.ColaReady, func(i, j int) bool {
		return global.ColaReady[i].EstimacionRafaga < global.ColaReady[j].EstimacionRafaga
	})

	proceso := global.ColaReady[0]
	global.ColaReady = global.ColaReady[1:]

	return proceso
}

func EvaluarDesalojo(nuevo Proceso) {
	if global.ConfigKernel.SchedulerAlgorithm != "SRTF" {
		return
	}

	ejecutando := global.ProcesoEnEjecucion

	if ejecutando.PCB.PID != 0 && nuevo.EstimacionRafaga < ejecutando.EstimacionRafaga {
		global.LoggerKernel.Log(fmt.Sprintf("Desalojando proceso %d por nuevo proceso %d", ejecutando.PCB.PID, nuevo.PCB.PID), log.INFO)
		EnviarInterrupcion(ejecutando.PCB.PID)
	}
}

func RecalcularRafaga(proceso *Proceso, rafagaReal float64) {
	alpha := global.ConfigKernel.Alpha // [0,1]
	proceso.EstimacionRafaga = alpha*rafagaReal + (1-alpha)*proceso.EstimacionRafaga
}
