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
	NEW    			string = "NEW"
	READY  			string = "READY"
	STOP   			string = "STOP"
	RUN    			string = "RUN"
	EXIT   			string = "EXIT"
	SUSP_READY 		string = "SUSP READY"
	SUSP_BLOCKED 	string = "SUSP BLOCKED"
)

var estado = []string{
	NEW,
	READY,
	STOP,
	RUN,
	EXIT,
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
		ArchivoPseudo: archivoPseudoCodigo,
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
				p := &global.ColaExit[0]
				FinalizarProceso(p)
				global.ColaExit = global.ColaExit[1:]

				// Intentar inicializar proceso desde NEW
				if intentarInicializarProceso(global.ColaNew) {
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
						break
					}
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IntentarInicializarDesdeNew() {
	if len(global.ColaNew) == 0 {
		return
	}

	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso := global.ColaNew[0]
		if SolicitarMemoria(proceso.MemoriaRequerida) {
			ActualizarEstadoPCB(&proceso.PCB, "Ready")
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaNew = global.ColaNew[1:] // Eliminar de NEW
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (FIFO)", proceso.PCB.PID), log.INFO)
		}

	case "CHICO":
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) {
				ActualizarEstadoPCB(&proceso.PCB, "Ready")
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (CHICO)", proceso.PCB.PID), log.INFO)
			} else {
				nuevaCola = append(nuevaCola, proceso) // No se movió, sigue en NEW
			}
		}
		global.ColaNew = nuevaCola // Actualiza la cola NEW con los procesos no movidos
	}
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

	// En cualquier otro caso, retornamos false (error)
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

func FinalizarProceso(p *Proceso) {
	ActualizarEstadoPCB(&p.PCB, EXIT)

	// Informar a Memoria (como vimos antes)
	err := InformarFinAMemoria(p.PID)
	if err != nil {
		// Manejo de error
		return
	}

	// Loguear metricas
	LoguearMetricas(p)

	// Eliminar de la cola
	global.ColaExecuting = filtrarCola(global.ColaExecuting,p)

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

/*	func EnviarProcessDataAMemoria(proceso Proceso, archPseudo string){
		pid := proceso.PCB.PID
		pseudoCodigo := proceso
		return 
	} */
