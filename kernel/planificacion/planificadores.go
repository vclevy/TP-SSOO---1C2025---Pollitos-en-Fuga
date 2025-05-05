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
	utilskernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	//estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
)

const (
	NEW          string = "NEW"
	READY        string = "READY"
	EXEC         string = "EXEC"
	EXIT         string = "EXIT"
	BLOCKED      string = "BLOCKED"
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
		EstimacionRafaga: float64(global.ConfigKernel.InitialEstimate), //? chequear
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
			// 1. Prioridad absoluta a procesos SUSP_READY
			if len(global.ColaSuspReady) > 0 {
				if IntentarCargarDesdeSuspReady() {
					continue
				}
			}

			// 2. Finalización de procesos (liberación de recursos)
			if len(global.ColaExit) > 0 {
				p := global.ColaExit[0]
				FinalizarProceso(&p)
				global.ColaExit = global.ColaExit[1:]

				// Al liberar recursos, intentar cargar SUSP_READY de nuevo
				continue
			}

			// 3. Carga de nuevos procesos (solo si no hay SUSP_READY esperando)
			if len(global.ColaNew) > 0 && len(global.ColaSuspReady) == 0 {
				switch global.ConfigKernel.SchedulerAlgorithm {
				case "FIFO":
					proceso := global.ColaNew[0]
					if SolicitarMemoria(proceso.MemoriaRequerida) {
						global.ColaNew = global.ColaNew[1:]
						ActualizarEstadoPCB(&proceso.PCB, READY)
						global.ColaReady = append(global.ColaReady, proceso)
						global.LoggerKernel.Log(fmt.Sprintf("## (%d) Pasa del estado NEW al estado READY", proceso.PCB.PID), log.INFO)
						EvaluarDesalojo(proceso)
					}
				case "CHICO":
					ordenada := make([]global.Proceso, len(global.ColaNew))
					copy(ordenada, global.ColaNew)
					sort.Slice(ordenada, func(i, j int) bool {
						return ordenada[i].MemoriaRequerida < ordenada[j].MemoriaRequerida
					})

					for _, proc := range ordenada {
						if SolicitarMemoria(proc.MemoriaRequerida) {
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
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IntentarCargarDesdeSuspReady() bool {
	for i := 0; i < len(global.ColaSuspReady); i++ {
		proceso := global.ColaSuspReady[i]

		if SolicitarMemoria(proceso.MemoriaRequerida) {
			if err := MoverAMemoria(proceso.PID); err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a memoria: %v", proceso.PID, err), log.ERROR)
				continue
			}

			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaSuspReady = append(global.ColaSuspReady[:i], global.ColaSuspReady[i+1:]...)

			global.LoggerKernel.Log(fmt.Sprintf("Proceso %d movido de SUSP_READY a READY", proceso.PID), log.INFO)
			return true
		}
	}
	return false
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
	endpoint := "verificarEspacioDisponible"
	url := fmt.Sprintf("http://%s:%d/%s?tamanio=%d", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory, endpoint, tamanio)

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
	url := "http://" + global.ConfigKernel.IPMemory + ":" + strconv.Itoa(global.ConfigKernel.Port_Memory) + "/finalizarProceso"

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
		global.LoggerKernel.Log(fmt.Sprintf("Error al informar finalización del proceso %d a Memoria: %s", p.PID, err.Error()), log.ERROR)
		return
	}

	LoguearMetricas(p) // (log obligatorio)
	global.ColaExecuting = utilskernel.FiltrarCola(global.ColaExecuting, p)
	global.ColaExit = append(global.ColaExit, p)

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
	// 1. Verificar CPUs disponibles
	if !HayCPUDisponible() {
		return
	}

	// 2. Intentar cargar procesos SUSP_READY si no hay READY
	if len(global.ColaReady) == 0 && len(global.ColaSuspReady) > 0 {
		IntentarCargarDesdeSuspReady()
		if len(global.ColaReady) == 0 {
			return // No se pudo cargar ningún proceso
		}
	}

	// 3. Selección de proceso según algoritmo
	var proceso global.Proceso
	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso = global.ColaReady[0]
		global.ColaReady = global.ColaReady[1:]
	case "SJF":
		proceso = seleccionarProcesoSJF(false)
	case "SRTF":
		proceso = seleccionarProcesoSJF(true)
	}

	// 4. Despachar proceso
	ActualizarEstadoPCB(&proceso.PCB, EXEC)
	global.ColaExecuting = append(global.ColaExecuting, proceso)
	global.LoggerKernel.Log(fmt.Sprintf("Proceso %d despachado a EXEC", proceso.PID), log.INFO)
}

func seleccionarProcesoSJF(desalojo bool) global.Proceso {
	if len(global.ColaReady) == 0 {
		return global.Proceso{}
	}

	copiaReady := make([]global.Proceso, len(global.ColaReady))
	copy(copiaReady, global.ColaReady)

	sort.Slice(copiaReady, func(i, j int) bool {
		return copiaReady[i].EstimacionRafaga < copiaReady[j].EstimacionRafaga
	})

	proceso := copiaReady[0]

	// Remover de la cola original
	for i, p := range global.ColaReady {
		if p.PID == proceso.PID {
			global.ColaReady = append(global.ColaReady[:i], global.ColaReady[i+1:]...)
			break
		}
	}

	return proceso
}

func EvaluarDesalojo(nuevo Proceso) {
	if global.ConfigKernel.SchedulerAlgorithm != "SRTF" {
		return
	}

	if len(global.ColaExecuting) == 0 {
		return
	}

	// Buscar el proceso en ejecución con mayor estimación de ráfaga
	procesoADesalojar := global.ColaExecuting[0]
	for _, proceso := range global.ColaExecuting {
		if proceso.EstimacionRafaga > procesoADesalojar.EstimacionRafaga {
			procesoADesalojar = proceso
		}
	}

	// Comparar el mejor candidato a desalojar contra el nuevo
	if nuevo.EstimacionRafaga < procesoADesalojar.EstimacionRafaga {
		global.LoggerKernel.Log(fmt.Sprintf("Desalojando proceso %d por nuevo proceso %d", procesoADesalojar.PCB.PID, nuevo.PCB.PID), log.INFO)
		//TODO EnviarInterrupcion(procesoADesalojar.PCB.PID) !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	}
}

func RecalcularRafaga(proceso *Proceso, rafagaReal float64) {
	alpha := global.ConfigKernel.Alpha // [0,1]
	proceso.EstimacionRafaga = alpha*rafagaReal + (1-alpha)*proceso.EstimacionRafaga
}

func HayCPUDisponible() bool {
	return global.CantidadCPUsOcupadas < global.CantidadCPUsTotales
}

func IniciarPlanificadorMedioPlazo() {
	go func() {
		for {
			// 1. Verificar procesos BLOCKED para suspensión
			for i := 0; i < len(global.ColaBlocked); i++ {
				proceso := &global.ColaBlocked[i]

				if time.Since(proceso.PCB.InicioEstado) > time.Duration(global.ConfigKernel.TiempoMaxBlocked)*time.Millisecond {
					suspenderProceso(proceso, i)
					i-- // Ajustar índice después de remover
				}
			}

			// 2. Verificar recursos para cargar procesos suspendidos
			if len(global.ColaSuspReady) > 0 {
				IntentarCargarDesdeSuspReady()
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func suspenderProceso(proceso *global.Proceso, index int) {
	// Cambiar estado
	ActualizarEstadoPCB(&proceso.PCB, SUSP_BLOCKED)

	// Mover a swap
	if err := MoverASwap(proceso.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a swap: %v", proceso.PID, err), log.ERROR)
		return
	}

	// Mover a cola SUSP_BLOCKED
	global.ColaSuspBlocked = append(global.ColaSuspBlocked, *proceso)
	global.ColaBlocked = append(global.ColaBlocked[:index], global.ColaBlocked[index+1:]...)

	global.LoggerKernel.Log(fmt.Sprintf("Proceso %d movido a SUSP_BLOCKED", proceso.PID), log.INFO)
}

func MoverASwap(pid int) error {
	url := fmt.Sprintf("http://%s:%d/moverASwap?pid=%d",
		global.ConfigKernel.IPMemory,
		global.ConfigKernel.Port_Memory,
		pid)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor de memoria")
	}

	return nil
}

func MoverAMemoria(pid int) error {
	url := fmt.Sprintf("http://%s:%d/moverAMemoria?pid=%d",
		global.ConfigKernel.IPMemory,
		global.ConfigKernel.Port_Memory,
		pid)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor de memoria")
	}

	return nil
}
