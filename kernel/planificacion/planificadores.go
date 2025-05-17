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

 func CrearProceso(tamanio int, archivoPseudoCodigo string) *Proceso {
	pcb := global.NuevoPCB()
	ActualizarEstadoPCB(pcb, NEW)

	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
		ArchivoPseudo:    archivoPseudoCodigo,
		EstimacionRafaga: float64(global.ConfigKernel.InitialEstimate),
	}

	global.LoggerKernel.Log(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID), log.INFO)
	global.MutexNew.Lock()
	global.ColaNew = append(global.ColaNew, &proceso)
	global.MutexNew.Unlock()
	return &proceso
}

func ActualizarEstadoPCB(pcb *PCB, nuevoEstado string) {
	ahora := time.Now()
	// Si ya tenía un estado previo, calculamos tiempo en ese estado
	if pcb.UltimoEstado != "" {
		duracion := int(ahora.Sub(pcb.InicioEstado).Milliseconds())
		pcb.MT[pcb.UltimoEstado] += duracion
	}
	// Log antes de actualizar el último estado
	global.LoggerKernel.Log(
		fmt.Sprintf("## (%d) Pasa del estado %s al estado %s", pcb.PID, pcb.UltimoEstado, nuevoEstado),
		log.INFO,
	)
	// Aumenta contador de veces en el nuevo estado
	pcb.ME[nuevoEstado] += 1
	// Actualiza último estado y momento de entrada
	pcb.UltimoEstado = nuevoEstado
	pcb.InicioEstado = ahora
}

func IniciarPlanificadorLargoPlazo() {
	go func() {
		<-global.InicioPlanificacionLargoPlazo
		global.LoggerKernel.Log("Iniciando planificación de largo plazo...", log.INFO)

		for {
			global.MutexSuspReady.Lock()
			if len(global.ColaSuspReady) > 0 {
				global.MutexSuspReady.Unlock()
				if IntentarCargarDesdeSuspReady() {
					continue
				}
			} else {
				global.MutexSuspReady.Unlock()
			}

			global.MutexExit.Lock()
			if len(global.ColaExit) > 0 {
				p := global.ColaExit[0]
				global.ColaExit = global.ColaExit[1:]
				global.MutexExit.Unlock()
				FinalizarProceso(p)
				continue
			}
			global.MutexExit.Unlock()

			global.MutexNew.Lock()
			colaNewLen := len(global.ColaNew)
			global.MutexNew.Unlock()

			global.MutexSuspReady.Lock()
			colaSuspReadyLen := len(global.ColaSuspReady)
			global.MutexSuspReady.Unlock()

			if colaNewLen > 0 && colaSuspReadyLen == 0 {
				switch global.ConfigKernel.SchedulerAlgorithm {
				case "FIFO":
					global.MutexNew.Lock()
					proceso := global.ColaNew[0]
					global.MutexNew.Unlock()
					if SolicitarMemoria(proceso.MemoriaRequerida) {
						global.MutexNew.Lock()
						global.ColaNew = global.ColaNew[1:]
						global.MutexNew.Unlock()
						ActualizarEstadoPCB(&proceso.PCB, READY)

						global.MutexReady.Lock()
						global.ColaReady = append(global.ColaReady, proceso)
						global.MutexReady.Unlock()
						EvaluarDesalojo(*proceso)
					}
				case "CHICO":
					global.MutexNew.Lock()
					ordenada := make([]*global.Proceso, len(global.ColaNew))
					copy(ordenada, global.ColaNew)
					global.MutexNew.Unlock()

					sort.Slice(ordenada, func(i, j int) bool {
						return ordenada[i].MemoriaRequerida < ordenada[j].MemoriaRequerida
					})

					for _, proc := range ordenada {
						if SolicitarMemoria(proc.MemoriaRequerida) {
							global.MutexNew.Lock()
							for i, p := range global.ColaNew {
								if p.PCB.PID == proc.PCB.PID {
									global.ColaNew = append(global.ColaNew[:i], global.ColaNew[i+1:]...)
									break
								}
							}
							global.MutexNew.Unlock()

							ActualizarEstadoPCB(&proc.PCB, READY)
							global.MutexReady.Lock()
							global.ColaReady = append(global.ColaReady, proc)
							global.MutexReady.Unlock()
							EvaluarDesalojo(*proc)
							break
						}
					}
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IniciarPlanificadorCortoPlazo() {
	go func() {
		for {
			if !HayCPUDisponible() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			global.MutexReady.Lock()
			readyLen := len(global.ColaReady)
			global.MutexReady.Unlock()
			global.MutexSuspReady.Lock()
			suspReadyLen := len(global.ColaSuspReady)
			global.MutexSuspReady.Unlock()

			if readyLen == 0 && suspReadyLen > 0 {
				IntentarCargarDesdeSuspReady()

				global.MutexReady.Lock()
				readyLen = len(global.ColaReady)
				global.MutexReady.Unlock()

				if readyLen == 0 {
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}

			var proceso *global.Proceso

			global.MutexReady.Lock()
			switch global.ConfigKernel.SchedulerAlgorithm {
			case "FIFO":
				proceso = global.ColaReady[0]
				global.ColaReady = global.ColaReady[1:]
			case "SJF":
				proceso = seleccionarProcesoSJF(false)
			case "SRTF":
				proceso = seleccionarProcesoSJF(true)
			}
			global.MutexReady.Unlock()

			if proceso == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			
			ActualizarEstadoPCB(&proceso.PCB, EXEC)
			global.MutexExecuting.Lock()
			global.ColaExecuting = append(global.ColaExecuting, proceso)
			global.MutexExecuting.Unlock()

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IniciarPlanificadorMedioPlazo() {
	go func() {
		for {
			nuevaColaBlocked := make([]*global.Proceso, 0)

			global.MutexBlocked.Lock()
			for _, proceso := range global.ColaBlocked {
				if time.Since(proceso.PCB.InicioEstado) > time.Duration(global.ConfigKernel.SuspensionTime)*time.Millisecond {
					global.MutexBlocked.Unlock() // liberar antes de suspender
					suspenderProceso(proceso)
					global.MutexBlocked.Lock()   // volver a bloquear
				} else {
					nuevaColaBlocked = append(nuevaColaBlocked, proceso)
				}
			}
			global.ColaBlocked = nuevaColaBlocked
			global.MutexBlocked.Unlock()

			global.MutexSuspReady.Lock()
			haySusp := len(global.ColaSuspReady) > 0
			global.MutexSuspReady.Unlock()
			if haySusp {
				IntentarCargarDesdeSuspReady()
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}


func IntentarCargarDesdeSuspReady() bool {
	global.MutexSuspReady.Lock()
	defer global.MutexSuspReady.Unlock()
	for i := 0; i < len(global.ColaSuspReady); i++ {
		proceso := global.ColaSuspReady[i]
		if SolicitarMemoria(proceso.MemoriaRequerida) {
			if err := MoverAMemoria(proceso.PID); err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a memoria: %v", proceso.PID, err), log.ERROR)
				continue
			}
			global.MutexReady.Lock()
			global.ColaReady = append(global.ColaReady, proceso)
			global.MutexReady.Unlock()
			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.ColaSuspReady = append(global.ColaSuspReady[:i], global.ColaSuspReady[i+1:]...)
			return true
		}
	}
	return false
}

func IntentarInicializarDesdeNew() bool {
	global.MutexNew.Lock()
	defer global.MutexNew.Unlock()

	if len(global.ColaNew) == 0 {
		return false
	}

	moved := false

	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso := global.ColaNew[0]
		if SolicitarMemoria(proceso.MemoriaRequerida) {
			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.MutexReady.Lock()
			global.ColaReady = append(global.ColaReady, proceso)
			global.MutexReady.Unlock()
			global.ColaNew = global.ColaNew[1:]
			moved = true
		}

	case "CHICO":
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []*Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) {
				ActualizarEstadoPCB(&proceso.PCB, READY)
				global.MutexReady.Lock()
				global.ColaReady = append(global.ColaReady, proceso)
				global.MutexReady.Unlock()
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
 	return respuesta.StatusCode == http.StatusOK
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

	LoguearMetricas(p)

	//TODO Liberar el PCB e intentar inicializar el siguiente en susp o new

	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()
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


func seleccionarProcesoSJF(_ bool) *global.Proceso {
	global.MutexReady.Lock()
	defer global.MutexReady.Unlock()

	if len(global.ColaReady) == 0 {
		return nil
	}

	copiaReady := make([]*global.Proceso, len(global.ColaReady))
	copy(copiaReady, global.ColaReady)

	sort.Slice(copiaReady, func(i, j int) bool {
		return copiaReady[i].EstimacionRafaga < copiaReady[j].EstimacionRafaga
	})

	proceso := copiaReady[0]

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
	global.MutexExecuting.Lock()
	if len(global.ColaExecuting) == 0 {
		return
	}
	
	procesoADesalojar := global.ColaExecuting[0]

	for _, proceso := range global.ColaExecuting {
		if proceso.EstimacionRafaga > procesoADesalojar.EstimacionRafaga {
			procesoADesalojar = proceso
		}
	}
	global.MutexExecuting.Unlock()
// 	// Comparar el mejor candidato a desalojar contra el nuevo
 	if nuevo.EstimacionRafaga < procesoADesalojar.EstimacionRafaga {
 		global.LoggerKernel.Log(fmt.Sprintf("Desalojando proceso %d por nuevo proceso %d", procesoADesalojar.PCB.PID, nuevo.PCB.PID), log.INFO)
 		//TODO EnviarInterrupcion(procesoADesalojar.PCB.PID) !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 	}
 }

func RecalcularRafaga(proceso *Proceso, rafagaReal float64) {
	alpha := global.ConfigKernel.Alpha 
	proceso.EstimacionRafaga = alpha*rafagaReal + (1-alpha)*proceso.EstimacionRafaga
}

func HayCPUDisponible() bool {
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			return true
		}
	}
	return false
}

func suspenderProceso(proceso *global.Proceso) {
	ActualizarEstadoPCB(&proceso.PCB, SUSP_BLOCKED)

	if err := MoverASwap(proceso.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a swap: %v", proceso.PID, err), log.ERROR)
		return
	}

	global.MutexSuspBlocked.Lock()
	global.ColaSuspBlocked = append(global.ColaSuspBlocked, proceso)
	global.MutexSuspBlocked.Unlock()

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