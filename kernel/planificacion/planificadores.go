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
		select {
		case <-global.NotifySuspReady:
			for {
				global.MutexSuspReady.Lock()
				tieneSuspReady := len(global.ColaSuspReady) > 0
				global.MutexSuspReady.Unlock()

				if !tieneSuspReady || !IntentarCargarDesdeSuspReady() {
					break
				}
			}

		case <-global.NotifyNew:
			global.MutexSuspReady.Lock()
			suspReadyVacio := len(global.ColaSuspReady) == 0
			global.MutexSuspReady.Unlock()

			if !suspReadyVacio {
				continue
			}

			global.MutexNew.Lock()
			if len(global.ColaNew) == 0 {
				global.MutexNew.Unlock()
				continue
			}
			global.MutexNew.Unlock()

			switch global.ConfigKernel.ReadyIngressAlgorithm {
			case "FIFO":
				global.MutexNew.Lock()
				if len(global.ColaNew) == 0 {
					global.MutexNew.Unlock()
					continue
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
						global.EliminarProcesoDeCola(&global.ColaNew, proc.PID)
						global.MutexNew.Unlock()

						ActualizarEstadoPCB(&proc.PCB, READY)
						global.AgregarAReady(proc)
						break
					}
				}
			}
		}
	}
}

func IniciarPlanificadorCortoPlazo() {
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

			case "SJF", "SRTF":
				nuevoProceso = seleccionarProcesoSJF(global.ConfigKernel.SchedulerAlgorithm == "SRTF") 
			}
			global.MutexReady.Unlock()

			if nuevoProceso == nil {
				break
			}

			if global.ConfigKernel.SchedulerAlgorithm == "SRTF" {
				if evaluarDesalojoSRTF(nuevoProceso) {
					global.LoggerKernel.Log(fmt.Sprintf("Proceso PID %d no asignado por desalojo SRTF", nuevoProceso.PCB.PID), log.DEBUG)
					continue
				}
				if AsignarCPU(nuevoProceso) {
					break
				}
			} else {
				if AsignarCPU(nuevoProceso) {
					break
				}
			}
		}
	}
}

func seleccionarProcesoSJF(usandoRestante bool) *global.Proceso {
	if len(global.ColaReady) == 0 {
		return nil
	}

	sort.SliceStable(global.ColaReady, func(i, j int) bool {
		var valI, valJ float64
		if usandoRestante {
			valI = EstimacionRestante(global.ColaReady[i])
			valJ = EstimacionRestante(global.ColaReady[j])
		} else {
			valI = float64(global.ColaReady[i].EstimacionRafaga)
			valJ = float64(global.ColaReady[j].EstimacionRafaga)
		}
		return valI < valJ
	})

	proceso := global.ColaReady[0] // ya no lo eliminás acá
	//global.LoggerKernel.Log(fmt.Sprintf("Longitud de Cola Ready: %d", len(global.ColaReady)), log.INFO)

	return proceso
}

func evaluarDesalojoSRTF(nuevoProceso *global.Proceso) bool {
	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()

	if utilskernel.HayCPUDisponible() || len(global.ColaExecuting) == 0 {
		return false
	}

	ejecutando := global.ColaExecuting[0]
	restanteEjecutando := EstimacionRestante(ejecutando)
	restanteNuevo := EstimacionRestante(nuevoProceso)

	if restanteNuevo < restanteEjecutando {
		cpu := utilskernel.BuscarCPUPorPID(ejecutando.PCB.PID)
		if cpu != nil {
			global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Desalojado por algoritmo SJF/SRT", ejecutando.PCB.PID), log.INFO)
			err := utilskernel.EnviarInterrupcionCPU(cpu, ejecutando.PCB.PID, ejecutando.PCB.PC)
			if err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción a CPU %s para proceso %d: %v", cpu.ID, ejecutando.PCB.PID, err), log.ERROR)
			}
			utilskernel.SacarProcesoDeCPU(ejecutando.PCB.PID)
		} else {
			global.LoggerKernel.Log(fmt.Sprintf("No se encontró CPU ejecutando proceso %d para interrupción", ejecutando.PCB.PID), log.ERROR)
		}
		return true
	}

	return false
}

func AsignarCPU(proceso *global.Proceso) bool {
	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("[ERROR] No se puede asignar PID %d, ya está finalizado", proceso.PID), log.ERROR)
		return false
	}

	global.MutexCPUs.Lock()
	var cpuLibre *global.CPU
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			cpuLibre = cpu
			break
		}
	}
	global.MutexCPUs.Unlock()

	if cpuLibre == nil {
		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
		return false
	}

	global.MutexReady.Lock()
	global.EliminarProcesoDeCola(&global.ColaReady, proceso.PID)
	global.MutexReady.Unlock()

	global.MutexCPUs.Lock()
	cpuLibre.ProcesoEjecutando = &proceso.PCB
	global.MutexCPUs.Unlock()

	if proceso.PCB.UltimoEstado != EXEC {

		proceso.MutexPCB.Lock()
		ActualizarEstadoPCB(&proceso.PCB, EXEC)
		proceso.MutexPCB.Unlock()
		global.AgregarAExecuting(proceso)

		go func(cpu *global.CPU, proceso *global.Proceso) {
			err := utilskernel.EnviarADispatch(cpu, proceso.PCB.PID, proceso.PCB.PC)
			if err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("[ERROR] Error en dispatch de proceso %d a CPU %s: %v", proceso.PID, cpu.ID, err), log.ERROR)

				global.MutexCPUs.Lock()
				cpu.ProcesoEjecutando = nil
				global.MutexCPUs.Unlock()

				if proceso.PCB.UltimoEstado != EXIT {
					global.LoggerKernel.Log(fmt.Sprintf("[TRACE] Reencolando proceso PID %d en READY tras error de dispatch", proceso.PID), log.DEBUG)
					global.AgregarAReady(proceso)
				}
			}
		}(cpuLibre, proceso)
	}

	return true
}

func ManejarDevolucionDeCPU(resp estructuras.RespuestaCPU) {
	var proceso *global.Proceso
	// Liberar CPU que ejecutaba este proceso
	global.LoggerKernel.Log(fmt.Sprintf("[DEBUG] Kernel recibe proceso PID %d con PC %d", resp.PID, resp.PC), log.DEBUG)

	global.MutexExecuting.Lock()
	for _, p := range global.ColaExecuting {
		if p.PCB.PID == resp.PID {
			proceso = p
			break
		}
	}
	global.MutexExecuting.Unlock()

	if proceso == nil {
		global.LoggerKernel.Log(fmt.Sprintf("[WARN] Proceso %d no encontrado en EXECUTING al devolver", resp.PID), log.DEBUG)
		return
	}

	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("[WARN] Se recibió devolución de CPU para PID %d pero ya estaba en EXIT", proceso.PID), log.DEBUG)
		return
	}

	proceso.MutexPCB.Lock()
	proceso.TiempoEjecutado += resp.RafagaReal
	proceso.PCB.PC = resp.PC
	proceso.MutexPCB.Unlock()
	RecalcularRafaga(proceso, resp.RafagaReal)
	global.LoggerKernel.Log(fmt.Sprintf("[DEBUG] Asignando a CPU proceso PID %d con PC %d", proceso.PID, proceso.PC), log.DEBUG)

	switch resp.Motivo {
	case "EXIT":
		utilskernel.SacarProcesoDeCPU(proceso.PID)
		FinalizarProceso(proceso)

	case "IO":
		utilskernel.SacarProcesoDeCPU(proceso.PID)

	case "READY":
		utilskernel.SacarProcesoDeCPU(proceso.PID)
		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()
		proceso.MutexPCB.Lock()
		ActualizarEstadoPCB(&proceso.PCB, READY)
		proceso.MutexPCB.Unlock()
		global.AgregarAReady(proceso)

	case "DUMP":
		utilskernel.SacarProcesoDeCPU(proceso.PID)
		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()
		proceso.MutexPCB.Lock()
		ActualizarEstadoPCB(&proceso.PCB, BLOCKED)
		proceso.MutexPCB.Unlock()
		global.AgregarABlocked(proceso)
	}

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

		global.ColaBlocked = nuevaColaBlocked
		global.MutexBlocked.Unlock()

		for _, p := range procesosASuspender {
			suspenderProceso(p)
			
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

			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.AgregarAReady(proceso)
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

	_ = IntentarCargarDesdeSuspReady()

	// 2. Siempre notificar al planificador de largo plazo para intentar NEW
	select {
	case global.NotifyNew <- struct{}{}:
	default:
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
	// Verificar que el proceso aún siga en BLOCKED antes de continuar
	if proceso.PCB.UltimoEstado != BLOCKED {
		return
	}

	global.MutexBlocked.Lock()
	global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
	global.MutexBlocked.Unlock()

	ActualizarEstadoPCB(&proceso.PCB, SUSP_BLOCKED)

	global.AgregarASuspBlocked(proceso)

	go func(pid int) {
		if err := utilskernel.MoverASwap(pid); err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error moviendo proceso %d a swap: %v", pid, err), log.ERROR)
		}
	}(proceso.PID)

	global.LoggerKernel.Log(fmt.Sprintf("Proceso %d suspendido y movido a SUSP_BLOCKED", proceso.PID), log.INFO)

	select {
    case global.NotifySuspReady <- struct{}{}:
    default: // si ya había señal pendiente, no bloquear
    }
}

func EstimacionRestante(p *Proceso) float64 {
	restante := p.EstimacionRafaga - p.TiempoEjecutado
	if restante < 0 {
		return 0
	}
	return restante
}