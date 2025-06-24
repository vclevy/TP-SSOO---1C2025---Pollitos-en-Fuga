package planificacion

import (
	"fmt"
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilskernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
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

	if pcb.UltimoEstado != "" {
		duracion := int(ahora.Sub(pcb.InicioEstado).Milliseconds())
		pcb.MT[pcb.UltimoEstado] += duracion
	}

	if nuevoEstado != NEW {
		global.LoggerKernel.Log(
			fmt.Sprintf("## (%d) Pasa del estado %s al estado %s", pcb.PID, pcb.UltimoEstado, nuevoEstado),
			log.INFO,
		)
	}

	pcb.ME[nuevoEstado] += 1
	pcb.UltimoEstado = nuevoEstado
	pcb.InicioEstado = ahora
}

func IniciarPlanificadorLargoPlazo() {
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
				global.MutexSuspReady.Lock()
				suspReadyVacio := len(global.ColaSuspReady) == 0
				global.MutexSuspReady.Unlock()

				if !suspReadyVacio {
					break
				}

				switch global.ConfigKernel.ReadyIngressALgorithm {
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
					}

				case "PMCP":
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
		}
	}
}

func IniciarPlanificadorCortoPlazo() {
	for {
		<-global.NotifyReady
		global.LoggerKernel.Log("Se recibió notificación de procesos en READY", log.DEBUG)

		for {
			if !utilskernel.HayCPUDisponible() && global.ConfigKernel.SchedulerAlgorithm != "SRTF" {
				global.LoggerKernel.Log("No hay CPU disponible y el algoritmo no es SRTF. Se detiene iteración.", log.DEBUG)
				break
			}

			global.MutexReady.Lock()
			if len(global.ColaReady) == 0 {
				global.MutexReady.Unlock()
				global.LoggerKernel.Log("Cola READY vacía. Se detiene iteración.", log.DEBUG)
				break
			}

			var nuevoProceso *global.Proceso
			switch global.ConfigKernel.SchedulerAlgorithm {
			case "FIFO":
				nuevoProceso = global.ColaReady[0]
				global.ColaReady = global.ColaReady[1:]
				global.LoggerKernel.Log(fmt.Sprintf("Seleccionado proceso FIFO PID %d", nuevoProceso.PCB.PID), log.DEBUG)

			case "SJF", "SRTF":
				nuevoProceso = seleccionarProcesoSJF(global.ConfigKernel.SchedulerAlgorithm == "SRTF")
				if nuevoProceso != nil {
					global.LoggerKernel.Log(fmt.Sprintf("Seleccionado proceso %s PID %d", global.ConfigKernel.SchedulerAlgorithm, nuevoProceso.PCB.PID), log.DEBUG)
				}
			}
			global.MutexReady.Unlock()

			if nuevoProceso == nil {
				global.LoggerKernel.Log("No se seleccionó ningún proceso. Se detiene iteración.", log.DEBUG)
				break
			}

			if global.ConfigKernel.SchedulerAlgorithm == "SRTF" {
				if evaluarDesalojoSRTF(nuevoProceso) {
					global.LoggerKernel.Log(fmt.Sprintf("Proceso PID %d no asignado por desalojo SRTF", nuevoProceso.PCB.PID), log.DEBUG)
					continue
				} else {
					global.LoggerKernel.Log(fmt.Sprintf("Proceso PID %d no ejecuta aún, vuelve a READY", nuevoProceso.PCB.PID), log.DEBUG)
					global.AgregarAReady(nuevoProceso)
					break
				}
			}

			global.LoggerKernel.Log(fmt.Sprintf("Asignando proceso PID %d a CPU", nuevoProceso.PCB.PID), log.DEBUG)
			AsignarCPU(nuevoProceso)
		}
	}
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

func ManejarDevolucionDeCPU(resp estructuras.RespuestaCPU) {
	var proceso *global.Proceso

	// Liberar CPU que ejecutaba este proceso
	global.MutexCPUs.Lock()
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando != nil && cpu.ProcesoEjecutando.PID == resp.PID {
			cpu.ProcesoEjecutando = nil
			break
		}
	}
	global.MutexCPUs.Unlock()

	// Buscar proceso sin removerlo todavía
	global.MutexExecuting.Lock()
	for _, p := range global.ColaExecuting {
		if p.PCB.PID == resp.PID {
			proceso = p
			break
		}
	}
	global.MutexExecuting.Unlock()

	if proceso == nil {
		global.LoggerKernel.Log(fmt.Sprintf("Proceso %d no encontrado en EXECUTING al devolver", resp.PID), log.DEBUG)
		return
	}

	proceso.PCB.PC = resp.PC
	RecalcularRafaga(proceso, resp.RafagaReal)

	switch resp.Motivo {
	case "EXIT":
		FinalizarProceso(proceso)

	case "BLOCKED":
		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()

		ActualizarEstadoPCB(&proceso.PCB, BLOCKED)
		global.AgregarABlocked(proceso)

	case "READY":
		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()

		ActualizarEstadoPCB(&proceso.PCB, READY)
		global.AgregarAReady(proceso)

		// Notificar al planificador solo si está esperando
		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
	}

	// Si el proceso no fue a READY, igual hay que notificar al planificador
	if resp.Motivo != "READY" {
		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
	}
}
func IniciarPlanificadorMedioPlazo() {
	for {
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

		global.LoggerKernel.Log(fmt.Sprintf("Procesos a suspender: %d", len(procesosASuspender)), log.DEBUG)
		global.ColaBlocked = nuevaColaBlocked
		global.MutexBlocked.Unlock()

		for _, p := range procesosASuspender {
			global.LoggerKernel.Log(fmt.Sprintf("Suspendiendo proceso PID %d", p.PCB.PID), log.DEBUG)
			suspenderProceso(p)

			select {
			case global.NotifySuspReady <- struct{}{}:
				global.LoggerKernel.Log("Notificando a largo plazo por SuspReady", log.DEBUG)
			default:
				global.LoggerKernel.Log("Notificación a largo plazo omitida (canal ya lleno)", log.DEBUG)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
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

	global.AgregarAExit(p)

	global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Finaliza el proceso", p.PID), log.INFO)

	LoguearMetricas(p)
	liberarPCB(p)

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
	msg := fmt.Sprintf("## (%d) - Métricas de estado:", p.PID)
	for _, unEstado := range estado {
		count := p.ME[unEstado]
		tiempo := p.MT[unEstado]
		msg += fmt.Sprintf(" %s (%d) (%d),", unEstado, count, tiempo)
	}

	msg = msg[:len(msg)-1]

	global.LoggerKernel.Log(msg, log.INFO)
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
