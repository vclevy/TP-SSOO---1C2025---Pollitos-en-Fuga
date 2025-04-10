package main

func main() {
// TODO: LEER ARCHIVO DE CONFIGURACION 
// TODO: LEVANTAR SERVIDOR
}

type PCB struct {
	PID      int            // Identificador unico del proceso
	PC       int            // Program Counter, tiene la direccion de la proxima instruccion a ejecutarse
	ME       map[string]int // Metricas de Estado: cantidad de veces en cada estado
	MT       map[string]int // Metricas de Tiempo por Estado: tiempo total por estado
}

func NuevoPCB(pid int) *PCB { //*PCB indica lo que te retorna
	return &PCB{ //&PCB indica la direccion de memoria
		PID: pid, //pasas por parametro solo el identificador del proceso, el resto arranca en "null"
		PC:  0,
		ME:  make(map[string]int), //make te reserva espacio en memoria, para inicializar estructuras como map o slice
		MT:  make(map[string]int),
	}
}

// Simula el paso por un estado
func (pcb *PCB) PasarPorEstado(estado string, duracionMs int) {
	pcb.ME[estado]++
	pcb.MT[estado] += duracionMs
}
