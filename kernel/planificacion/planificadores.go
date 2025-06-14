package planificacion

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilskernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"sort"
	"time"
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
	global.AgregarANew(&proceso)
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
			// Prioridad: SuspReady
			for {
				global.MutexSuspReady.Lock()
				tieneSuspReady := len(global.ColaSuspReady) > 0
				global.MutexSuspReady.Unlock()

				if !tieneSuspReady || !IntentarCargarDesdeSuspReady() {
					break
				}
			}

			select {
			case <-global.NotifyNew:
				global.MutexNew.Lock()
				colaNewLen := len(global.ColaNew)
				global.MutexNew.Unlock()

				if colaNewLen > 0 {
					switch global.ConfigKernel.SchedulerAlgorithm {
					case "FIFO":
						global.MutexNew.Lock()
						if len(global.ColaNew) == 0 {
							global.MutexNew.Unlock()
							break
						}
						proceso := global.ColaNew[0]
						global.MutexNew.Unlock()

						if utilskernel.InicializarProceso(proceso) {
							global.MutexNew.Lock()
							global.ColaNew = global.ColaNew[1:]
							global.MutexNew.Unlock()

							ActualizarEstadoPCB(&proceso.PCB, READY)
							global.AgregarAReady(proceso)
							EvaluarDesalojo(proceso)
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
							if utilskernel.InicializarProceso(proc) {
								global.MutexNew.Lock()
								for i, p := range global.ColaNew {
									if p.PCB.PID == proc.PCB.PID {
										global.ColaNew = append(global.ColaNew[:i], global.ColaNew[i+1:]...)
										break
									}
								}
								global.MutexNew.Unlock()

								ActualizarEstadoPCB(&proc.PCB, READY)
								global.AgregarAReady(proc)
								EvaluarDesalojo(proc)
								break
							}
						}
					}
				}

			case <-time.After(1 * time.Second):
				global.MutexNew.Lock()
				colaNewLen := len(global.ColaNew)
				global.MutexNew.Unlock()

				if colaNewLen > 0 {
					select {
					case global.NotifyNew <- struct{}{}:
					default:
					}
				}

				global.MutexExit.Lock()
				if len(global.ColaExit) > 0 {
					p := global.ColaExit[0]
					global.ColaExit = global.ColaExit[1:]
					global.MutexExit.Unlock()
					FinalizarProceso(p)
				} else {
					global.MutexExit.Unlock()
				}
			}
		}
	}()
}


func IniciarPlanificadorCortoPlazo() {
	go func() {
		for {
			<-global.NotifyReady

			for {
				if !utilskernel.HayCPUDisponible() && global.ConfigKernel.SchedulerAlgorithm != "SRTF" {
					break
				}

				global.MutexReady.Lock()
				if len(global.ColaReady) == 0 {
					global.MutexReady.Unlock()
					break
				}

				var nuevoProceso *global.Proceso
				switch global.ConfigKernel.SchedulerAlgorithm {
				case "FIFO":
					nuevoProceso = global.ColaReady[0]
					global.ColaReady = global.ColaReady[1:]
				case "SJF", "SRTF":
					nuevoProceso = seleccionarProcesoSJF(global.ConfigKernel.SchedulerAlgorithm == "SRTF")
				}
				global.MutexReady.Unlock()

				if nuevoProceso == nil {
					break
				}

				if global.ConfigKernel.SchedulerAlgorithm == "SRTF" {
					if evaluarDesalojoSRTF(nuevoProceso) {
						continue
					} else {
						// vuelve a ready pq no se ejecuta todavía
						global.AgregarAReady(nuevoProceso)
						break
					}
				}

				AsignarCPU(nuevoProceso)
			}
		}
	}()
}

func evaluarDesalojoSRTF(nuevoProceso *global.Proceso) bool {
	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()

	if len(global.ColaExecuting) == 0 {
		return false
	}

	ejecutando := global.ColaExecuting[0]
	restanteEjecutando := EstimacionRestante(ejecutando)
	restanteNuevo := EstimacionRestante(nuevoProceso)

	if restanteNuevo < restanteEjecutando {
		cpu := utilskernel.BuscarCPUPorPID(ejecutando.PCB.PID)
		if cpu != nil {
			err := utilskernel.EnviarInterrupcionCPU(cpu, ejecutando.PCB.PID, ejecutando.PCB.PC)
			if err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción a CPU %s para proceso %d: %v", cpu.ID, ejecutando.PCB.PID, err), log.ERROR)
			}
		} else {
			global.LoggerKernel.Log(fmt.Sprintf("No se encontró CPU ejecutando proceso %d para interrupción", ejecutando.PCB.PID), log.ERROR)
		}
		return true
	}
	return false
}

func AsignarCPU(proceso *global.Proceso) {
	global.MutexCPUs.Lock()

	var cpuLibre *global.CPU
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			cpuLibre = cpu
			break
		}
	}

	if cpuLibre != nil {
		cpuLibre.ProcesoEjecutando = &proceso.PCB
	}
	global.MutexCPUs.Unlock()

	if cpuLibre == nil {
		global.LoggerKernel.Log(fmt.Sprintf("No hay CPU disponible para proceso %d, vuelve a READY", proceso.PID), log.INFO)
		global.AgregarAReady(proceso)
		return
	}

	ActualizarEstadoPCB(&proceso.PCB, EXEC)
	global.AgregarAExecuting(proceso)

	go func(cpu *global.CPU, proceso *global.Proceso) {
		err := utilskernel.EnviarADispatch(cpu, proceso.PCB.PID, proceso.PCB.PC)
		if err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error en dispatch de proceso %d a CPU %s: %v", proceso.PID, cpu.ID, err), log.ERROR)

			global.MutexCPUs.Lock()
			cpu.ProcesoEjecutando = nil
			global.MutexCPUs.Unlock()

			global.AgregarAReady(proceso)
			return
		}
	}(cpuLibre, proceso)
}

func ManejarDevolucionDeCPU(pid int, nuevoPC int, motivo string, rafagaReal float64) {
	var proceso *global.Proceso

	// Liberar CPU que ejecutaba este proceso
	global.MutexCPUs.Lock()
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando != nil && cpu.ProcesoEjecutando.PID == pid {
			cpu.ProcesoEjecutando = nil
			break
		}
	}
	global.MutexCPUs.Unlock()

	global.MutexExecuting.Lock()
	for i, p := range global.ColaExecuting {
		if p.PCB.PID == pid {
			proceso = p
			global.ColaExecuting = append(global.ColaExecuting[:i], global.ColaExecuting[i+1:]... )
			break
		}
	}
	global.MutexExecuting.Unlock()

	if proceso == nil {
		global.LoggerKernel.Log(fmt.Sprintf("Proceso %d no encontrado en EXECUTING al devolver", pid), log.DEBUG)
		return
	}

	proceso.PCB.PC = nuevoPC
	RecalcularRafaga(proceso, rafagaReal)

	switch motivo {
	case "EXIT":
		FinalizarProceso(proceso)

	case "BLOCKED":
		ActualizarEstadoPCB(&proceso.PCB, BLOCKED)
		global.AgregarABlocked(proceso)

	case "READY":
		ActualizarEstadoPCB(&proceso.PCB, READY)
		global.AgregarAReady(proceso)

		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
	}
}

func IniciarPlanificadorMedioPlazo() {
	go func() {
		for {
			// Recolectar procesos a suspender
			global.MutexBlocked.Lock()
			var nuevaColaBlocked []*global.Proceso
			var procesosASuspender []*global.Proceso

			for _, proceso := range global.ColaBlocked {
				tiempoBloqueado := time.Since(proceso.PCB.InicioEstado)
				if tiempoBloqueado > time.Duration(global.ConfigKernel.SuspensionTime)*time.Millisecond {
					procesosASuspender = append(procesosASuspender, proceso)
				} else {
					nuevaColaBlocked = append(nuevaColaBlocked, proceso)
				}
			}
			global.ColaBlocked = nuevaColaBlocked
			global.MutexBlocked.Unlock()

			// Suspender fuera del lock
			for _, p := range procesosASuspender {
				suspenderProceso(p)

				// Notificar al planificador largo plazo para intentar cargar desde SuspReady
				select {
				case global.NotifySuspReady <- struct{}{}:
				default:
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func IntentarCargarDesdeSuspReady() bool {
	global.MutexSuspReady.Lock()
	defer global.MutexSuspReady.Unlock()

	nuevaCola := make([]*global.Proceso, 0, len(global.ColaSuspReady))
	cambio := false

	for _, proceso := range global.ColaSuspReady {
		if cambio {
			nuevaCola = append(nuevaCola, proceso)
			continue
		}

		if utilskernel.VerificarEspacioDisponible(proceso.MemoriaRequerida) {
			if err := utilskernel.MoverAMemoria(proceso.PID); err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a memoria: %v", proceso.PID, err), log.ERROR)
				nuevaCola = append(nuevaCola, proceso)
				continue
			}

			global.AgregarAReady(proceso)
			ActualizarEstadoPCB(&proceso.PCB, READY)
			cambio = true
		} else {
			nuevaCola = append(nuevaCola, proceso)
		}
	}

	global.ColaSuspReady = nuevaCola
	return cambio
}

func FinalizarProceso(p *Proceso) {
	ActualizarEstadoPCB(&p.PCB, EXIT)

	if err := utilskernel.InformarFinAMemoria(p.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error al informar finalización del proceso %d a Memoria: %s", p.PID, err.Error()), log.ERROR)
	}

	global.MutexExecuting.Lock()
	global.EliminarProcesoDeCola(&global.ColaExecuting, p.PID)
	global.MutexExecuting.Unlock()

	global.MutexExit.Lock()
	global.AgregarAExit(p)
	global.MutexExit.Unlock()

	liberarPCB(p)
	LoguearMetricas(p)

	// Intentar cargar desde SuspReady primero
	if !IntentarCargarDesdeSuspReady() {
		// Si no había ninguno para mover desde SuspReady, notificar NEW
		select {
		case global.NotifyNew <- struct{}{}:
		default:
		}
	}
}

func liberarPCB(p *Proceso) {
	if p == nil {
		return
	}

	for k := range p.ME {
		delete(p.ME, k)
	}
	for k := range p.MT {
		delete(p.MT, k)
	}

	p.PC = 0
	p.UltimoEstado = ""
	p.InicioEstado = time.Time{}
	p.MemoriaRequerida = 0
	p.ArchivoPseudo = ""
	p.EstimacionRafaga = 0
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

func seleccionarProcesoSJF(usandoRestante bool) *global.Proceso {
	global.MutexReady.Lock()
	defer global.MutexReady.Unlock()

	if len(global.ColaReady) == 0 {
		return nil
	}

	sort.SliceStable(global.ColaReady, func(i, j int) bool {
		if usandoRestante {
			return EstimacionRestante(global.ColaReady[i]) < EstimacionRestante(global.ColaReady[j])
		}
		return global.ColaReady[i].EstimacionRafaga < global.ColaReady[j].EstimacionRafaga
	})

	proceso := global.ColaReady[0]
	global.ColaReady = global.ColaReady[1:]

	return proceso
}

func EvaluarDesalojo(nuevo *global.Proceso) {
	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()

	if len(global.ColaExecuting) == 0 {
		return
	}

	var procesoADesalojar *global.Proceso
	maxRafaga := -1.0

	for _, p := range global.ColaExecuting {
		if p.EstimacionRafaga > maxRafaga {
			maxRafaga = p.EstimacionRafaga
			procesoADesalojar = p
		}
	}

	if procesoADesalojar != nil && nuevo.EstimacionRafaga < maxRafaga {
		global.LoggerKernel.Log(fmt.Sprintf("Desalojando proceso %d por nuevo proceso %d", procesoADesalojar.PCB.PID, nuevo.PCB.PID), log.INFO)
		cpu := utilskernel.BuscarCPUPorPID(procesoADesalojar.PCB.PID)
		if cpu == nil {
			global.LoggerKernel.Log(fmt.Sprintf("No se encontró CPU ejecutando proceso %d para interrupción", procesoADesalojar.PCB.PID), log.ERROR)
			return
		}
		if err := utilskernel.EnviarInterrupcionCPU(cpu, procesoADesalojar.PCB.PID, procesoADesalojar.PCB.PC); err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción: %v", err), log.ERROR)
		}
	}
}

func RecalcularRafaga(proceso *Proceso, rafagaReal float64) {
	alpha := global.ConfigKernel.Alpha
	proceso.EstimacionRafaga = alpha*rafagaReal + (1-alpha)*proceso.EstimacionRafaga
}

func suspenderProceso(proceso *global.Proceso) {
	if err := utilskernel.MoverASwap(proceso.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a swap: %v", proceso.PID, err), log.ERROR)
		return
	}

	ActualizarEstadoPCB(&proceso.PCB, SUSP_BLOCKED)
	global.AgregarASuspBlocked(proceso)
}

func EstimacionRestante(p *Proceso) float64 {
	restante := p.EstimacionRafaga - p.TiempoEjecutado
	if restante < 0 {
		return 0
	}
	return restante
}
