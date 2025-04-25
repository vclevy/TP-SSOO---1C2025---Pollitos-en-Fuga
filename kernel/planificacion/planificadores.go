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

func CrearProceso(tamanio int) Proceso {
	pcb := global.NuevoPCB()
	ActualizarEstadoPCB(pcb, NEW)

	proceso := Proceso{
		PCB:              *pcb,
		MemoriaRequerida: tamanio,
	}

	global.LoggerKernel.Log(fmt.Sprintf("## (%d) Se crea el proceso - Estado: NEW", pcb.PID), log.INFO) //! LOG OBLIGATORIO: Creacion de Proceso
	return proceso

}

func PlanificarProcesoLargoPlazo(pseudoCodigo string, proceso Proceso) {
	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		if len(global.ColaNew) == 0 {
			if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				//TODO PasarPseudocodigoAMemoria(proceso)
				ActualizarEstadoPCB(&proceso.PCB, READY)
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("## (%d) Pasa del estado"+NEW+" al estado"+ READY, proceso.PCB.PID), log.INFO)  //! LOG OBLIGATORIO: Cambio de Estado
				return
			}
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (FIFO)", proceso.PCB.PID), log.DEBUG)

	case "CHICO":
		if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
			//TODO PasarPseudocodigoAMemoria(proceso)
			ActualizarEstadoPCB(&proceso.PCB, "Ready")
			global.ColaReady = append(global.ColaReady, proceso)
			global.LoggerKernel.Log(fmt.Sprintf("## (%d) Pasa del estado"+NEW+" al estado"+ READY, proceso.PCB.PID), log.INFO)  //! LOG OBLIGATORIO: Cambio de Estado
			return
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (CHICO)", proceso.PCB.PID), log.DEBUG)
	}
}


func IntentarInicializarDesdeNew() {
	if len(global.ColaNew) == 0 {
		return
	}

	switch global.ConfigKernel.SchedulerAlgorithm {
	case "FIFO":
		proceso := global.ColaNew[0]
		if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
			//TODO PasarPseudocodigoAMemoria(proceso)
			ActualizarEstadoPCB(&proceso.PCB, "Ready")
			global.ColaReady = append(global.ColaReady, proceso)
			global.ColaNew = global.ColaNew[1:]
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (FIFO)", proceso.PID), log.INFO)
		}

	case "CHICO":
		sort.Slice(global.ColaNew, func(i, j int) bool {
			return global.ColaNew[i].MemoriaRequerida < global.ColaNew[j].MemoriaRequerida
		})
		nuevaCola := []Proceso{}
		for _, proceso := range global.ColaNew {
			if SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				//TODO PasarPseudocodigoAMemoria(proceso)
				ActualizarEstadoPCB(&proceso.PCB, "Ready")
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d movido de NEW a READY (CHICO)", proceso.PID), log.INFO)
			} else {
				nuevaCola = append(nuevaCola, proceso)
			}
		}
		global.ColaNew = nuevaCola
	}
}

func SolicitarMemoria(tamanio int) int {
	cliente := &http.Client{}
	endpoint := "tamanioProceso/" + strconv.Itoa(tamanio)//Supongo que tamanioProceso es el handler de Memoria (Fran)
	url := fmt.Sprintf("http://%s:%d/%s", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return http.StatusInternalServerError
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return http.StatusInternalServerError
	}
	defer respuesta.Body.Close()

	return respuesta.StatusCode
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

	// Loguear métricas
	LoguearMetricas(p)

	// Eliminar de la cola
	EliminarProcesoDeCola(global.ColaExecuting,p)

	// Intentar iniciar otro
	
}

func EliminarProcesoDeCola(cola []global.Proceso, p *Proceso) {
	cola = filtrarCola(cola,p)
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

	
	// func NuevoPlanificador(algo AlgoritmoPlanificacion, enviarAMemoria func(*Proceso) bool) *PlanificadorLargoPlazo {
	// 	return &PlanificadorLargoPlazo{
	// 		ColaNew:        []*procesos.Proceso{},
	// 		Algoritmo:      algo,
	// 		Estado:         STOP,
	// 		EnviarAMemoria: enviarAMemoria,
	// 	}
	// }
	// 
	// func (p *PlanificadorLargoPlazo) EsperarInicio() {
	// 	fmt.Println("Planificador en STOP. Presiona ENTER para iniciar...")
	// 	bufio.NewReader(os.Stdin).ReadBytes('\n')
	// 	p.Estado = NEW
	// 	fmt.Println("Planificador iniciado.")
	// }
	// 
	// func (p *PlanificadorLargoPlazo) AgregarProceso(proceso *Proceso) {
	// 	p.ColaNew = append(p.ColaNew, proceso)
	// 	p.ordenarCola()
	// 
	// 	if len(p.ColaNew) == 1 { //? No entendi xq el == 1 y no >=1
	// 		p.intentarInicializar()
	// 	}
	// }
	// 
	// func (p *PlanificadorLargoPlazo) ordenarCola() {
	// 	switch p.Algoritmo {
	// 	case FIFO:
	// 		// No hace falta ordenar, se mantiene el orden
	// 	case MasChicoPrimero:
	// 		sort.Slice(p.ColaNew, func(i, j int) bool {
	// 			return p.ColaNew[i].MemoriaRequerida < p.ColaNew[j].MemoriaRequerida
	// 		})
	// 	}
	// }
	// 
	// func (p *PlanificadorLargoPlazo) intentarInicializar() {
	// 	for len(p.ColaNew) > 0 {
	// 		proceso := p.ColaNew[0]
	// 		if p.EnviarAMemoria(proceso) {
	// 			fmt.Printf("Proceso %d movido a READY.\n", proceso.PID)
	// 			p.ColaNew = p.ColaNew[1:]
	// 		} else {
	// 			fmt.Printf("Memoria no disponible para proceso %d. Esperando...\n", proceso.PID)
	// 			break
	// 		}
	// 	}
	// }