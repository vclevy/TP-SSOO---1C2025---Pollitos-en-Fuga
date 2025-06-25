package planificacion

import (
	"fmt"
	"sort"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilskernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
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

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorOrange = "\033[38;5;208m" // naranja aproximado usando color 256
)

type PCB = global.PCB
type Proceso = global.Proceso

func CrearProceso(tamanio int, archivoPseudoCodigo string) *Proceso {
	pcb := global.NuevoPCB()

	global.MutexUltimoPID.Lock()
	global.UltimoPID++
	global.MutexUltimoPID.Unlock()

	ActualizarEstadoPCB(pcb, NEW)

	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
		ArchivoPseudo:    archivoPseudoCodigo,
		EstimacionRafaga: float64(global.ConfigKernel.InitialEstimate),
	}

	global.LoggerKernel.Log(ColorYellow+fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID)+ColorReset, log.INFO)
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
	global.LoggerKernel.Log("Iniciando planificación de largo plazo...", log.DEBUG)

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

				switch global.ConfigKernel.ReadyIngressAlgorithm {
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
				nuevoProceso = seleccionarProcesoSJF(global.ConfigKernel.SchedulerAlgorithm == "SRTF") //esta ya elimina de ready
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
					global.LoggerKernel.Log(fmt.Sprintf("Asignando proceso PID %d a CPU", nuevoProceso.PCB.PID), log.DEBUG)
					AsignarCPU(nuevoProceso)
					break
				}
			}

			global.LoggerKernel.Log(fmt.Sprintf("Asignando proceso PID %d a CPU", nuevoProceso.PCB.PID), log.DEBUG)
			AsignarCPU(nuevoProceso)
		}
	}
}

func seleccionarProcesoSJF(usandoRestante bool) *global.Proceso {

	if len(global.ColaReady) == 0 {
		global.LoggerKernel.Log("seleccionarProcesoSJF: ColaReady vacía, retorna nil", log.DEBUG)
		return nil
	}

	sort.SliceStable(global.ColaReady, func(i, j int) bool {

		var valI, valJ float64
		if usandoRestante {
			valI = EstimacionRestante(global.ColaReady[i])
			global.LoggerKernel.Log(fmt.Sprintf("EstimacionRestante PID %d: %f", global.ColaReady[i].PCB.PID, valI), log.DEBUG)
			valJ = EstimacionRestante(global.ColaReady[j])
			global.LoggerKernel.Log(fmt.Sprintf("EstimacionRestante PID %d: %f", global.ColaReady[j].PCB.PID, valJ), log.DEBUG)
		} else {
			valI = float64(global.ColaReady[i].EstimacionRafaga)
			valJ = float64(global.ColaReady[j].EstimacionRafaga)
			global.LoggerKernel.Log(fmt.Sprintf("Comparando rafaga PID %d: %f vs PID %d: %f",
				global.ColaReady[i].PCB.PID, valI, global.ColaReady[j].PCB.PID, valJ), log.DEBUG)
		}

		return valI < valJ
	})

	proceso := global.ColaReady[0]
	global.ColaReady = global.ColaReady[1:]

	global.LoggerKernel.Log(fmt.Sprintf("seleccionarProcesoSJF: seleccionado PID %d", proceso.PCB.PID), log.DEBUG)
	return proceso
}

func evaluarDesalojoSRTF(nuevoProceso *global.Proceso) bool {
	global.LoggerKernel.Log(fmt.Sprintf("evaluarDesalojoSRTF: nuevo proceso PID %d", nuevoProceso.PCB.PID), log.DEBUG)
	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()

	if utilskernel.HayCPUDisponible() {
		return false
	}

	if len(global.ColaExecuting) == 0 {
		global.LoggerKernel.Log("evaluarDesalojoSRTF: no hay procesos ejecutando, no se desaloja", log.DEBUG)
		return false
	}

	ejecutando := global.ColaExecuting[0]
	restanteEjecutando := EstimacionRestante(ejecutando)
	restanteNuevo := EstimacionRestante(nuevoProceso)

	global.LoggerKernel.Log(fmt.Sprintf(
		"evaluarDesalojoSRTF: ejecutando PID %d restante %f, nuevo PID %d restante %f",
		ejecutando.PCB.PID, restanteEjecutando, nuevoProceso.PCB.PID, restanteNuevo), log.DEBUG)

	if restanteNuevo < restanteEjecutando {
		cpu := utilskernel.BuscarCPUPorPID(ejecutando.PCB.PID)
		if cpu != nil {
			global.LoggerKernel.Log(fmt.Sprintf("evaluarDesalojoSRTF: enviando interrupción a CPU %s para proceso %d", cpu.ID, ejecutando.PCB.PID), log.DEBUG)
			err := utilskernel.EnviarInterrupcionCPU(cpu, ejecutando.PCB.PID, ejecutando.PCB.PC)
			if err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción a CPU %s para proceso %d: %v", cpu.ID, ejecutando.PCB.PID, err), log.ERROR)
			}
		} else {
			global.LoggerKernel.Log(fmt.Sprintf("No se encontró CPU ejecutando proceso %d para interrupción", ejecutando.PCB.PID), log.ERROR)
		}
		return true
	}

	global.LoggerKernel.Log("evaluarDesalojoSRTF: no se desaloja, nuevo proceso no tiene menor restante", log.DEBUG)
	return false
}

func AsignarCPU(proceso *global.Proceso) {
	//No intentar asignar si ya está finalizado
	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("[ERROR] No se puede asignar PID %d, ya está finalizado", proceso.PID), log.ERROR)
		return
	}

	global.LoggerKernel.Log(fmt.Sprintf("Intentando asignar CPU al proceso PID %d", proceso.PID), log.DEBUG)

	global.MutexCPUs.Lock()

	var cpuLibre *global.CPU
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			cpuLibre = cpu
			break
		}
	}

	if cpuLibre != nil {
		global.LoggerKernel.Log(fmt.Sprintf("CPU libre encontrada: %s para proceso %d", cpuLibre.ID, proceso.PID), log.DEBUG)
		cpuLibre.ProcesoEjecutando = &proceso.PCB
	} else {
		global.LoggerKernel.Log(fmt.Sprintf("No hay CPU disponible para proceso %d, vuelve a READY", proceso.PID), log.DEBUG)
		global.MutexCPUs.Unlock()

		// Evitar reencolar si ya está finalizado
		if proceso.PCB.UltimoEstado != EXIT {
			global.AgregarAReady(proceso)
		} else {
			global.LoggerKernel.Log(fmt.Sprintf("[WARN] Proceso %d finalizado no se reencola en READY", proceso.PID), log.ERROR)
		}
		return
	}

	global.MutexCPUs.Unlock()

	// No actualizar a EXEC si ya está en EXEC
	if proceso.PCB.UltimoEstado != EXEC {
		ActualizarEstadoPCB(&proceso.PCB, EXEC)
	}

	global.AgregarAExecuting(proceso)

	go func(cpu *global.CPU, proceso *global.Proceso) {
		global.LoggerKernel.Log(fmt.Sprintf("Enviando dispatch proceso %d a CPU %s", proceso.PID, cpu.ID), log.DEBUG)

		err := utilskernel.EnviarADispatch(cpu, proceso.PCB.PID, proceso.PCB.PC)
		if err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error en dispatch de proceso %d a CPU %s: %v", proceso.PID, cpu.ID, err), log.ERROR)

			global.MutexCPUs.Lock()
			cpu.ProcesoEjecutando = nil
			global.MutexCPUs.Unlock()

			// otra vez: no lo reencolamos si está finalizado
			if proceso.PCB.UltimoEstado != EXIT {
				global.AgregarAReady(proceso)
			}
			return
		}

		global.LoggerKernel.Log(fmt.Sprintf("Dispatch proceso %d a CPU %s exitoso", proceso.PID, cpu.ID), log.DEBUG)
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

	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("Se recibió devolución de CPU para PID %d pero ya estaba en EXIT", proceso.PID), log.DEBUG)
		return
	}

	proceso.TiempoEjecutado += resp.RafagaReal
	proceso.PCB.PC = resp.PC
	global.LoggerKernel.Log(fmt.Sprintf("Proceso PID %d - TiempoEjecutado actualizado a %f", proceso.PCB.PID, proceso.TiempoEjecutado), log.DEBUG)

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

		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
	}

	// Notificar al planificador si no se notificó antes (por BLOCKED o EXIT)
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

		//!global.LoggerKernel.Log(fmt.Sprintf("Procesos a suspender: %d", len(procesosASuspender)), log.DEBUG) dsps lo descomento
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

	global.MutexReady.Lock()
	global.EliminarProcesoDeCola(&global.ColaReady, p.PID)
	global.MutexReady.Unlock()

	global.MutexBlocked.Lock()
	global.EliminarProcesoDeCola(&global.ColaBlocked, p.PID)
	global.MutexBlocked.Unlock()

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

	global.MutexCPUs.Lock()
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando != nil && cpu.ProcesoEjecutando.PID == p.PID {
			cpu.ProcesoEjecutando = nil
		}
	}
	global.MutexCPUs.Unlock()
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

func RecalcularRafaga(proceso *Proceso, rafagaReal float64) {
	alpha := global.ConfigKernel.Alpha
	proceso.EstimacionRafaga = alpha*rafagaReal + (1-alpha)*proceso.EstimacionRafaga
}

func suspenderProceso(proceso *global.Proceso) {
	global.MutexBlocked.Lock()
	removido := global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
	global.MutexBlocked.Unlock()

	if !removido {
		global.LoggerKernel.Log(fmt.Sprintf("Advertencia: proceso %d no estaba en BLOCKED al suspender", proceso.PID), log.ERROR)
	}

	if err := utilskernel.MoverASwap(proceso.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a swap: %v", proceso.PID, err), log.ERROR)
		return
	}

	ActualizarEstadoPCB(&proceso.PCB, SUSP_BLOCKED)
	global.AgregarASuspBlocked(proceso)
	
	global.LoggerKernel.Log(fmt.Sprintf("Proceso %d suspendido y movido a SUSP_BLOCKED", proceso.PID), log.INFO)
}

func EstimacionRestante(p *Proceso) float64 {
	restante := p.EstimacionRafaga - p.TiempoEjecutado
	if restante < 0 {
		return 0
	}
	return restante
}
