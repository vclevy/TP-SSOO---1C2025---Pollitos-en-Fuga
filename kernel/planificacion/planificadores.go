package planificacion

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	
	procesos "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
)

type PCB = procesos.PCB
type Proceso = procesos.Proceso
type Estado string

const (
	NEW    			Estado = "NEW"
	READY  			Estado = "READY"
	STOP   			Estado = "STOP"
	RUN    			Estado = "RUN"
	EXIT   			Estado = "EXIT"
	SUSP_READY 		Estado = "SUSP READY"
	SUSP_BLOCKED 	Estado = "SUSP BLOCKED"
)


type AlgoritmoPlanificacion string

const (
	FIFO              AlgoritmoPlanificacion = "FIFO"
	MasChicoPrimero   AlgoritmoPlanificacion = "SJF"
)

type PlanificadorLargoPlazo struct {
	ColaNew        []*Proceso
	Algoritmo      AlgoritmoPlanificacion
	Estado         Estado
	EnviarAMemoria func(*Proceso) bool // Simula petición a Memoria
}

func NuevoPlanificador(algo AlgoritmoPlanificacion, enviarAMemoria func(*Proceso) bool) *PlanificadorLargoPlazo {
	return &PlanificadorLargoPlazo{
		ColaNew:        []*procesos.Proceso{},
		Algoritmo:      algo,
		Estado:         STOP,
		EnviarAMemoria: enviarAMemoria,
	}
}

func (p *PlanificadorLargoPlazo) EsperarInicio() {
	fmt.Println("Planificador en STOP. Presiona ENTER para iniciar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	p.Estado = NEW
	fmt.Println("Planificador iniciado.")
}

func (p *PlanificadorLargoPlazo) AgregarProceso(proceso *Proceso) {
	p.ColaNew = append(p.ColaNew, proceso)
	p.ordenarCola()

	if len(p.ColaNew) == 1 { //? No entendi xq el == 1 y no >=1
		p.intentarInicializar()
	}
}

func (p *PlanificadorLargoPlazo) ordenarCola() {
	switch p.Algoritmo {
	case FIFO:
		// No hace falta ordenar, se mantiene el orden
	case MasChicoPrimero:
		sort.Slice(p.ColaNew, func(i, j int) bool {
			return p.ColaNew[i].MemoriaRequerida < p.ColaNew[j].MemoriaRequerida
		})
	}
}

func (p *PlanificadorLargoPlazo) intentarInicializar() {
	for len(p.ColaNew) > 0 {
		proceso := p.ColaNew[0]
		if p.EnviarAMemoria(proceso) {
			fmt.Printf("Proceso %d movido a READY.\n", proceso.PID)
			p.ColaNew = p.ColaNew[1:]
		} else {
			fmt.Printf("Memoria no disponible para proceso %d. Esperando...\n", proceso.PID)
			break
		}
	}
}
