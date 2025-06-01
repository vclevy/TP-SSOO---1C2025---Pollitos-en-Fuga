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
	global.AgregarANew(&proceso)
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
			select {
			case <-global.NotifySuspReady:
				for {
					global.MutexSuspReady.Lock()
					tieneSuspReady := len(global.ColaSuspReady) > 0
					global.MutexSuspReady.Unlock()

					if !tieneSuspReady {
						break
					}
					if !IntentarCargarDesdeSuspReady() {
						break
					}
				}

			case <-global.NotifyNew:
				global.MutexSuspReady.Lock()
				colaSuspReadyLen := len(global.ColaSuspReady)
				global.MutexSuspReady.Unlock()

				if colaSuspReadyLen == 0 {
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

							if SolicitarMemoria(proceso.MemoriaRequerida) {
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
									global.AgregarAReady(proc)
									EvaluarDesalojo(proc)
									break
								}
							}
						}
					}
				}

			case <-time.After(1 * time.Second):
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
			<-global.NotifyReady // Esperar señal para planificar

			for {
				if !HayCPUDisponible() && global.ConfigKernel.SchedulerAlgorithm != "SRTF" {
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
				case "SJF":
					nuevoProceso = seleccionarProcesoSJF(false)
				case "SRTF":
					nuevoProceso = seleccionarProcesoSJF(true)
				}
				global.MutexReady.Unlock()

				if nuevoProceso == nil {
					break
				}

				// Para SRTF, se puede evaluar si desalojar antes de asignar CPU
				if global.ConfigKernel.SchedulerAlgorithm == "SRTF" {
					global.MutexExecuting.Lock()
					if len(global.ColaExecuting) > 0 {
						ejecutando := global.ColaExecuting[0]
						restanteEjecutando := EstimacionRestante(ejecutando)
						restanteNuevo := EstimacionRestante(nuevoProceso)

						if restanteNuevo < restanteEjecutando {
							// Desalojar el proceso que está en ejecución
							global.ColaExecuting = global.ColaExecuting[1:]
							global.MutexExecuting.Unlock()

							ActualizarEstadoPCB(&ejecutando.PCB, READY)
							global.AgregarAReady(ejecutando)
							//global.LoggerKernel.Log("Proceso %d desalojado por %d (SRTF)", ejecutando.PCB.PID, nuevoProceso.PCB.PID)
						} else {
							global.AgregarAReady(nuevoProceso)
							global.MutexExecuting.Unlock()
							break
						}
					} else {
						global.MutexExecuting.Unlock()
					}
				}

				// Asignar CPU: cambiar estado, mover a EXECUTING y notificar CPU
				AsignarCPU(nuevoProceso)
			}
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

func AsignarCPU(proceso *global.Proceso) {
	ActualizarEstadoPCB(&proceso.PCB, EXEC)

	global.MutexExecuting.Lock()
	global.ColaExecuting = append(global.ColaExecuting, proceso)
	global.MutexExecuting.Unlock()

	// Aquí debés llamar a tu módulo CPU para que corra el proceso
	// Ejemplo:
	err := EnviarA_CPU(proceso)
	if err != nil {
		global.LoggerKernel.Logf("Error enviando proceso %d a CPU: %v", proceso.PCB.PID, err)
		// Manejar error: sacar de EXEC, poner en READY o New según corresponda
	}
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

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error creando request para solicitar memoria: %v", err), log.ERROR)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(req)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error enviando request para solicitar memoria: %v", err), log.ERROR)
		return false
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		global.LoggerKernel.Log(fmt.Sprintf("Memoria respondió con status %d para solicitud de %d bytes", respuesta.StatusCode, tamanio), log.ERROR)
		return false
	}

	return true
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

	if err := InformarFinAMemoria(p.PID); err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error al informar finalización del proceso %d a Memoria: %s", p.PID, err.Error()), log.ERROR)
	}

	LoguearMetricas(p)

	liberarPCB(p)

	// Reordenar ejecución con locks mínimos
	global.MutexExecuting.Lock()
	global.ColaExecuting = utilskernel.FiltrarCola(global.ColaExecuting, p)
	global.MutexExecuting.Unlock()

	global.MutexExit.Lock()
	global.ColaExit = append(global.ColaExit, p)
	global.MutexExit.Unlock()

	// Intentar iniciar procesos en espera
	iniciarProcesosEnEspera()
}

func liberarPCB(p *Proceso) {
	if p == nil {
		return
	}

	// Limpiar mapas para liberar referencias internas
	for k := range p.ME {
		delete(p.ME, k)
	}
	for k := range p.MT {
		delete(p.MT, k)
	}

	// Limpiar datos de PCB y proceso si quieres (no obligatorio, GC los limpia)
	p.PC = 0
	p.UltimoEstado = ""
	p.InicioEstado = time.Time{}
	p.MemoriaRequerida = 0
	p.ArchivoPseudo = ""
	p.EstimacionRafaga = 0
}

func iniciarProcesosEnEspera() {
	global.MutexSuspReady.Lock()
	defer global.MutexSuspReady.Unlock()
	if len(global.ColaSuspReady) == 0 {
		return
	}
	p := global.ColaSuspReady[0]
	global.ColaSuspReady = global.ColaSuspReady[1:]

	global.MutexReady.Lock()
	global.ColaReady = append(global.ColaReady, p)
	global.MutexReady.Unlock()
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

func seleccionarProcesoSJF(preemptivo bool) *global.Proceso {
	global.MutexReady.Lock()
	defer global.MutexReady.Unlock()

	if len(global.ColaReady) == 0 {
		return nil
	}

	// Si es preemptivo, se selecciona el de menor rafaga que el ejecutando, sino el menor de la cola
	// Pero como no hay mutex para ejecutando aquí, dejamos la lógica simple como antes.

	copiaReady := make([]*global.Proceso, len(global.ColaReady))
	copy(copiaReady, global.ColaReady)

	sort.Slice(copiaReady, func(i, j int) bool {
		return copiaReady[i].EstimacionRafaga < copiaReady[j].EstimacionRafaga
	})

	proceso := copiaReady[0]

	// Remover de ColaReady
	for i, p := range global.ColaReady {
		if p.PID == proceso.PID {
			global.ColaReady = append(global.ColaReady[:i], global.ColaReady[i+1:]...)
			break
		}
	}

	return proceso
}


func EvaluarDesalojo(nuevo *Proceso) {
	// Bloqueamos mutex si tenés para evitar condiciones de carrera
	global.MutexExecuting.Lock()
	defer global.MutexExecuting.Unlock()

	if len(global.ColaExecuting) == 0 {
		return
	}

	var procesoADesalojar *Proceso
	maxRafaga := -1.0

	for _, p := range global.ColaExecuting {
		if p.EstimacionRafaga > maxRafaga {
			maxRafaga = p.EstimacionRafaga
			procesoADesalojar = p
		}
	}

	if procesoADesalojar != nil && nuevo.EstimacionRafaga < maxRafaga {
		global.LoggerKernel.Log(fmt.Sprintf("Desalojando proceso %d por nuevo proceso %d", procesoADesalojar.PCB.PID, nuevo.PCB.PID), log.INFO)
		if err := EnviarInterrupcion(procesoADesalojar.PCB.PID); err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción: %v", err), log.ERROR)
		}
	}
}


func EnviarInterrupcion(pid int) error {
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando != nil && cpu.ProcesoEjecutando.PID == pid {
			// Construimos URL del endpoint de interrupción, asumiendo puerto para interrupciones es cpu.Puerto
			url := fmt.Sprintf("http://%s:%d/interrumpir?pid=%d", cpu.IP, cpu.Puerto, pid)

			// Enviar POST sin cuerpo (o puede ser nil, si tu API lo permite)
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				return fmt.Errorf("error enviando interrupción a CPU %s: %v", cpu.ID, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("CPU %s devolvió estado %d al intentar interrumpir proceso %d", cpu.ID, resp.StatusCode, pid)
			}

			return nil
		}
	}
	return fmt.Errorf("no se encontró CPU ejecutando el proceso %d para interrumpir", pid)
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

func EstimacionRestante(p *Proceso) float64 {
    restante := p.EstimacionRafaga - p.TiempoEjecutado
    if restante < 0 {
        return 0
    }
    return restante
}
