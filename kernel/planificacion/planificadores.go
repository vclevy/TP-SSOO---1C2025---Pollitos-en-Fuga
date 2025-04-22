package planificacion

import (
	"fmt"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	procesos "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
)

type PCB = procesos.PCB
type Proceso = procesos.Proceso

const (
	NEW    			string = "NEW"
	READY  			string = "READY"
	STOP   			string = "STOP"
	RUN    			string = "RUN"
	EXIT   			string = "EXIT"
	SUSP_READY 		string = "SUSP READY"
	SUSP_BLOCKED 	string = "SUSP BLOCKED"
)

func PlanificarProcesoLargoPlazo(pseudoCodigo string, proceso Proceso) {
	switch global.AlgoritmoLargoPlazo {
	case "FIFO":
		if len(global.ColaNew) == 0 {
			if procesos.SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
				//TODO PasarPseudocodigoAMemoria(proceso)
				procesos.ActualizarEstadoPCB(&proceso.PCB, READY)
				global.ColaReady = append(global.ColaReady, proceso)
				global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", proceso.PCB.PID), log.INFO)
				return
			}
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (FIFO)", proceso.PCB.PID), log.INFO)

	case "CHICO":
		if procesos.SolicitarMemoria(proceso.MemoriaRequerida) == http.StatusOK {
			//TODO PasarPseudocodigoAMemoria(proceso)
			procesos.ActualizarEstadoPCB(&proceso.PCB, READY)
			global.ColaReady = append(global.ColaReady, proceso)
			global.LoggerKernel.Log(fmt.Sprintf("PID: %d pasó a READY", proceso.PCB.PID), log.INFO)
			return
		}
		global.ColaNew = append(global.ColaNew, proceso)
		global.LoggerKernel.Log(fmt.Sprintf("PID: %d encolado en NEW (CHICO)", proceso.PCB.PID), log.INFO)
	}
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
