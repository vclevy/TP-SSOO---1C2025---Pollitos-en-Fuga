package utilsKernel

import(
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"fmt"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	

)

func DispositivoExiste(nombre string) bool {
	_, ok := global.DispositivosIO[nombre]
	return ok
}

func BloquearProceso(pid int, dispositivo string) {
	proc := BuscarProcesoPorPID(pid)
	if proc != nil {
		planificacion.ActualizarEstadoPCB(&proc.PCB, global.BLOCKED)
		global.ColaBlocked = append(global.ColaBlocked, *proc)
		global.LoggerKernel.Log(fmt.Sprintf("PID %d bloqueado por IO %s", pid, dispositivo), logger.INFO)
	}
}

type EsperandoIO struct {
	PID      int
	Duracion int
}

var colaIO = make(map[string][]EsperandoIO)

func EncolarEnIO(nombre string, pid, duracion int) {
	colaIO[nombre] = append(colaIO[nombre], EsperandoIO{PID: pid, Duracion: duracion})
}

func DispositivoLibre(nombre string) bool {
	_, ocupado := global.IOOcupados[nombre]
	return !ocupado
}

func EnviarAIO(nombre string) {
	if len(colaIO[nombre]) == 0 {
		return
	}

	proximo := colaIO[nombre][0]
	colaIO[nombre] = colaIO[nombre][1:]

	url := fmt.Sprintf("http://%s:%d/io", global.DispositivosIO[nombre].IP, global.DispositivosIO[nombre].Puerto)
	payload := map[string]int{"pid": proximo.PID, "duracion": proximo.Duracion}
	data, _ := json.Marshal(payload)

	_, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err == nil {
		global.IOOcupados[nombre] = proximo.PID
		global.LoggerKernel.Log(fmt.Sprintf("PID %d enviado a dispositivo %s", proximo.PID, nombre), logger.INFO)
	}
}
