package planificacion

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilskernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"sort"
	"strconv"
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
			if !utilskernel.HayCPUDisponible() {
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
					// Se pidió un desalojo, esperamos a que se libere CPU
					global.LoggerKernel.Log(fmt.Sprintf("Se solicitó desalojo para asignar PID %d (SRTF)", nuevoProceso.PCB.PID), log.DEBUG)
					break 
				}
				global.LoggerKernel.Log(fmt.Sprintf("No hay tal desalojo para PID %d (SRTF)", nuevoProceso.PCB.PID), log.DEBUG)
			}

			asignado := AsignarCPU(nuevoProceso)
			if asignado {
				continue
			} else {
				break
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

	restanteNuevo := EstimacionRestante(nuevoProceso)
	var peorProceso *global.Proceso
	var peorRestante float64 = -1

	for _, ejecutando := range global.ColaExecuting {
		restante := EstimacionRestante(ejecutando)
		if restante > peorRestante {
			peorRestante = restante
			peorProceso = ejecutando
		}
	}

	global.LoggerKernel.Log(fmt.Sprintf(
		"[DEBUG] SRTF: nuevo PID %d (%.2f) vs peor ejecutando PID %d (%.2f)",
		nuevoProceso.PID, restanteNuevo, peorProceso.PCB.PID, peorRestante,
	), log.DEBUG)

	if restanteNuevo < peorRestante {
		cpu := utilskernel.BuscarCPUPorPID(peorProceso.PCB.PID)
		if cpu != nil {
			global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Desalojado por SRTF (nuevo PID %d)", peorProceso.PCB.PID, nuevoProceso.PID), log.INFO)
			err := utilskernel.EnviarInterrupcionCPU(cpu, peorProceso.PCB.PID, peorProceso.PCB.PC)
			if err != nil {
				global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción a CPU %s para proceso %d: %v", cpu.ID, peorProceso.PCB.PID, err), log.ERROR)
			}
			utilskernel.SacarProcesoDeCPU(peorProceso.PCB.PID)
			return true
		}
		global.LoggerKernel.Log(fmt.Sprintf("No se encontró CPU ejecutando proceso %d para interrupción", peorProceso.PCB.PID), log.ERROR)
	}

	return false
}
func AsignarCPU(proceso *global.Proceso) bool {
	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("No se puede asignar PID %d, ya está finalizado", proceso.PID), log.ERROR)
		return false
	}

	global.MutexCPUs.Lock()

	global.CPUsConectadas = append(global.CPUsConectadas[1:], global.CPUsConectadas[0])

	var cpuLibre *global.CPU
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			cpuLibre = cpu
			cpu.ProcesoEjecutando = &proceso.PCB
			break
		}
	}

	global.MutexCPUs.Unlock()

	if cpuLibre == nil {
		global.NotificarReady()
		return false
	}

	global.MutexReady.Lock()
	global.EliminarProcesoDeCola(&global.ColaReady, proceso.PID)
	global.MutexReady.Unlock()

	// ✅ Siempre agregar a EXEC y actualizar estado si es necesario
	if proceso.PCB.UltimoEstado != EXEC {
		ActualizarEstadoPCB(&proceso.PCB, EXEC)
	}
	global.AgregarAExecuting(proceso)

	err := utilskernel.EnviarADispatch(cpuLibre, proceso.PCB.PID, proceso.PCB.PC)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dispatch de proceso %d a CPU %s: %v", proceso.PID, cpuLibre.ID, err), log.ERROR)

		global.MutexCPUs.Lock()
		cpuLibre.ProcesoEjecutando = nil
		global.MutexCPUs.Unlock()

		if proceso.PCB.UltimoEstado != EXIT {
			global.LoggerKernel.Log(fmt.Sprintf("Reencolando proceso PID %d en READY tras error de dispatch", proceso.PID), log.DEBUG)
			global.AgregarAReady(proceso)
		}
		return false
	}

	return true
}


func ManejarDevolucionDeCPU(resp estructuras.RespuestaCPU) {
	var proceso *global.Proceso

	global.LoggerKernel.Log(fmt.Sprintf("Kernel recibe proceso PID %d con PC %d", resp.PID, resp.PC), log.DEBUG)

	global.MutexExecuting.Lock()
	global.MutexCPUs.Lock()

	for _, p := range global.ColaExecuting {
		if p.PCB.PID == resp.PID {
			proceso = p
			break
		}
	}

	global.MutexCPUs.Unlock()
	global.MutexExecuting.Unlock()

	if proceso == nil {
		global.LoggerKernel.Log(fmt.Sprintf("Proceso %d no encontrado en EXECUTING al devolver", resp.PID), log.DEBUG)
		return
	}

	if proceso.PCB.UltimoEstado == EXIT {
		global.LoggerKernel.Log(fmt.Sprintf("Se recibió devolución de CPU para PID %d pero ya estaba en EXIT", proceso.PID), log.DEBUG)
		return
	}

	proceso.PCB.PC = resp.PC

	// Guardamos estimación previa para calcular el restante después
	estimacionAnterior := proceso.EstimacionRafaga

	// ✅ Recalcular solo si terminó su ráfaga
	if resp.Motivo == "EXIT" || resp.Motivo == "IO" || resp.Motivo == "DUMP" {
		RecalcularRafaga(proceso, resp.RafagaReal)
	}

	// Acumulamos el tiempo ejecutado (siempre)
	proceso.TiempoEjecutado += resp.RafagaReal

	global.LoggerKernel.Log(
		fmt.Sprintf("PID %d - Ráfaga ejecutada: %.2f ms | Total ejecutado: %.2f ms",
			proceso.PID, resp.RafagaReal, proceso.TiempoEjecutado),
		log.DEBUG,
	)

	restante := estimacionAnterior - proceso.TiempoEjecutado
	if restante < 0 {
		restante = 0
	}
	global.LoggerKernel.Log(
		fmt.Sprintf("PID %d - Estimación restante: %.2f ms", proceso.PID, restante),
		log.DEBUG,
	)

	global.LoggerKernel.Log(fmt.Sprintf("[DEBUG] Asignando a CPU proceso PID %d con PC %d", proceso.PID, proceso.PC), log.DEBUG)

	switch resp.Motivo {
	case "EXIT":
		global.LoggerKernel.Log("## ("+strconv.Itoa(proceso.PID)+") - Solicitó syscall: <EXIT>", log.INFO)
		utilskernel.SacarProcesoDeCPU(proceso.PID)
		FinalizarProceso(proceso)

	case "IO":
		utilskernel.SacarProcesoDeCPU(proceso.PID)
		ManejarSolicitudIO(resp.PID, resp.IO.IoSolicitada, resp.IO.TiempoEstimado)

	case "READY":
		utilskernel.SacarProcesoDeCPU(proceso.PID)

		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()

		ActualizarEstadoPCB(&proceso.PCB, READY)
		global.AgregarAReady(proceso)

	case "DUMP":
		utilskernel.SacarProcesoDeCPU(proceso.PID)

		global.MutexExecuting.Lock()
		global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
		global.MutexExecuting.Unlock()

		ActualizarEstadoPCB(&proceso.PCB, BLOCKED)
		global.AgregarABlocked(proceso)

		global.LoggerKernel.Log("## ("+strconv.Itoa(proceso.PID)+") - Solicitó syscall: <DUMP_MEMORY>", log.INFO)

		err := utilskernel.SolicitarDumpAMemoria(proceso.PID)
		if err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", proceso.PID, err.Error()), log.ERROR)

			global.MutexBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
			global.MutexBlocked.Unlock()

			FinalizarProceso(proceso)
		} else {
			global.MutexBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
			global.MutexBlocked.Unlock()

			ActualizarEstadoPCB(&proceso.PCB, READY)
			global.AgregarAReady(proceso)
			global.LoggerKernel.Log("AGREGAR A READY (desde syscall DUMP)", log.DEBUG)
		}
	}

	global.NotificarReady()
}


func ManejarSolicitudIO(pid int, nombre string, tiempoUso int) error {
	global.LoggerKernel.Log("## ("+strconv.Itoa(pid)+") - Solicitó syscall: <IO>", log.INFO)

	global.IOListMutex.Lock()
	dispositivos := utilskernel.ObtenerDispositivoIO(nombre)
	global.IOListMutex.Unlock()

	global.MutexExecuting.Lock() 
	proceso := utilskernel.BuscarProcesoPorPID(global.ColaExecuting, pid)
	if proceso == nil {
		global.MutexExecuting.Unlock() 
		return fmt.Errorf("no se pudo obtener el proceso en EXECUTING (PID %d)", pid)
	}

	global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
	global.MutexExecuting.Unlock()

	if len(dispositivos) == 0 {
		global.LoggerKernel.Log(fmt.Sprintf("Dispositivo IO %s no existe, enviando %d a EXIT", nombre, pid), log.ERROR)
		FinalizarProceso(proceso)
		return fmt.Errorf("dispositivo IO %s no existe", nombre)
	}

	procesoEncolado := &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

	ActualizarEstadoPCB(&proceso.PCB, BLOCKED)
	global.AgregarABlocked(proceso)

	for _, dispositivo := range dispositivos {
		dispositivo.Mutex.Lock()
		if !dispositivo.Ocupado {
			dispositivo.Ocupado = true
			dispositivo.ProcesoEnUso = procesoEncolado
			dispositivo.Mutex.Unlock()

			global.LoggerKernel.Log("## ("+strconv.Itoa(pid)+") - Bloqueado por IO: <"+dispositivo.Nombre+">", log.INFO)
			go utilskernel.EnviarAIO(dispositivo, pid, tiempoUso)
			return nil
		}
		dispositivo.Mutex.Unlock()
	}

	// Si todos ocupados, encolar en el primero
	primero := dispositivos[0]
	primero.Mutex.Lock()
	primero.ColaEspera = append(primero.ColaEspera, procesoEncolado)
	global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Encolado en %s (Ocupado)", pid, primero.Nombre), log.DEBUG)
	primero.Mutex.Unlock()

	return nil
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

	// Siempre notificar al planificador de largo plazo para intentar NEW
	select {
	case global.NotifyNew <- struct{}{}:
	default:
	}

	utilskernel.SacarProcesoDeCPU(p.PID)
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
	proceso.TiempoEjecutado = 0 // RESET al recalcular estimación (fin de ráfaga)
}

func suspenderProceso(proceso *global.Proceso) {
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

	global.LoggerKernel.Log(fmt.Sprintf("Proceso %d suspendido y movido a SUSP_BLOCKED", proceso.PID), log.DEBUG)

	select {
	case global.NotifySuspReady <- struct{}{}:
	default: 
	}
}

func EstimacionRestante(p *Proceso) float64 {
	restante := p.EstimacionRafaga - p.TiempoEjecutado
	if restante < 0 {
		return 0
	}
	return restante
}
