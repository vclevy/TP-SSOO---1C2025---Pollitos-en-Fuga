package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO
type IODevice = global.IODevice
type PCB = planificacion.PCB
type SyscallIO = estructuras.SyscallIO

type Respuesta struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}

func RecibirPaquete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		global.LoggerKernel.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerKernel.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete PaqueteHandshakeIO
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerKernel.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}

	global.LoggerKernel.Log("Kernel recibió paquete desde IO - Nombre: "+paquete.NombreIO+" | IP IO: "+paquete.IPIO+" | Puerto Io: "+strconv.Itoa(paquete.PuertoIO), log.DEBUG)

	ioConectado := &IODevice{
		Nombre:       paquete.NombreIO,
		IP:           paquete.IPIO,
		Puerto:       paquete.PuertoIO,
		Ocupado:      false,
		ProcesoEnUso: nil,
		ColaEspera:   nil,
	}

	global.IOConectados = append(global.IOConectados, ioConectado)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func INIT_PROC(w http.ResponseWriter, r *http.Request) {
	//? archivo := r.URL.Query().Get("archivo") //? no se que hacer con el archivo este
	tamanioStr := r.URL.Query().Get("tamanio")

	pcb := global.NuevoPCB()

	tamanio, _ := strconv.Atoi(tamanioStr)

	procesoCreado := global.Proceso{PCB: *pcb, MemoriaRequerida: tamanio}
	global.LoggerKernel.Log(fmt.Sprintf("Proceso creado: %+v", procesoCreado), log.DEBUG)

	global.ColaNew = append(global.ColaNew, global.Proceso(procesoCreado)) // no estoy segura si esta bien la sintaxis
}

func HandshakeConCPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id") //* me gustaria usar body en vez de queryParam
	ip := r.URL.Query().Get("ip")
	puerto := r.URL.Query().Get("puerto")

	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", id, ip, puerto), log.DEBUG)

	w.WriteHeader(http.StatusOK)

	pcb := global.NuevoPCB()

	respuesta := map[string]interface{}{
		"pid": pcb.PID,
		"pc":  pcb.PC,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

func IO(w http.ResponseWriter, r *http.Request) {
	var syscall SyscallIO
	if err := json.NewDecoder(r.Body).Decode(&syscall); err != nil {
		http.Error(w, "Error al parsear la syscall", http.StatusBadRequest)
		return
	}
	nombre := syscall.IoSolicitada
	tiempoUso := syscall.TiempoEstimado
	pid := syscall.PIDproceso

	global.IOListMutex.RLock()
	dispositivos := utilsKernel.ObtenerDispositivoIO(nombre)
	global.IOListMutex.RUnlock()

	if dispositivos == nil || len(dispositivos) == 0 {
		proceso := BuscarProcesoPorPID(global.ColaExecuting, pid)
	
		if proceso != nil {
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXIT)
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Dispositivo %s no existe", nombre)
		return
	}

	proceso := BuscarProcesoPorPID(global.ColaExecuting, pid)
	if proceso == nil {
		http.Error(w, "No se pudo obtener el proceso actual", http.StatusInternalServerError)
		return
	}

	// Buscar un dispositivo libre
	for _, dispositivo := range dispositivos {
		dispositivo.Mutex.Lock()
		if !dispositivo.Ocupado {
			dispositivo.Ocupado = true
			dispositivo.ProcesoEnUso = proceso
			dispositivo.Mutex.Unlock()

			go utilsKernel.EnviarAIO(dispositivo, proceso.PCB.PID, tiempoUso)
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
			return
		}
		dispositivo.Mutex.Unlock()
	}

	// Si todos están ocupados, se encola en el primero
	primero := dispositivos[0]
	primero.Mutex.Lock()
	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
	primero.ColaEspera = append(primero.ColaEspera, proceso)
	primero.Mutex.Unlock()
}


func manejarIOOcupado(io *global.IODevice, proceso *global.Proceso, w http.ResponseWriter) {
	// Agregar a cola de espera
	io.ColaEspera = append(io.ColaEspera, proceso)

	// Cambiar estado según si está en memoria o swap
	if proceso.PCB.UltimoEstado == planificacion.SUSP_READY || proceso.PCB.UltimoEstado == planificacion.SUSP_BLOCKED {
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.SUSP_BLOCKED)
	} else {
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
		global.ColaBlocked = append(global.ColaBlocked, *proceso)
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Proceso %d en cola para %s", proceso.PID, io.Nombre)
}

func manejarIOLibre(io *global.IODevice, proceso *global.Proceso, tiempoUso int, w http.ResponseWriter) {
	// Asignar dispositivo
	io.Ocupado = true
	io.ProcesoEnUso = proceso

	// Ejecutar operación IO en goroutine
	go func() {
		time.Sleep(time.Duration(tiempoUso) * time.Millisecond)

		io.Mutex.Lock()
		defer io.Mutex.Unlock()

		// Liberar dispositivo
		io.Ocupado = false
		io.ProcesoEnUso = nil

		// Manejar cola de espera
		if len(io.ColaEspera) > 0 {
			siguiente := io.ColaEspera[0]
			io.ColaEspera = io.ColaEspera[1:]

			// Verificar si el proceso está suspendido
			for i, p := range global.ColaSuspBlocked {
				if p.PID == siguiente.PID {
					// Mover a SUSP_READY
					planificacion.ActualizarEstadoPCB(&p.PCB, planificacion.SUSP_READY)
					global.ColaSuspReady = append(global.ColaSuspReady, p)
					global.ColaSuspBlocked = append(global.ColaSuspBlocked[:i], global.ColaSuspBlocked[i+1:]...)
					return
				}
			}

			// Si no estaba suspendido, mover a READY
			planificacion.ActualizarEstadoPCB(&siguiente.PCB, planificacion.READY)
			global.ColaReady = append(global.ColaReady, *siguiente)

			// Remover de BLOCKED si estaba allí
			for i, p := range global.ColaBlocked {
				if p.PID == siguiente.PID {
					global.ColaBlocked = append(global.ColaBlocked[:i], global.ColaBlocked[i+1:]...)
					break
				}
			}
		}
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Proceso %d accediendo a %s por %d ms", proceso.PID, io.Nombre, tiempoUso)
}

func BuscarProcesoPorPID(cola []global.Proceso, pid int) (*global.Proceso) {
	for i := range cola {
		if cola[i].PCB.PID == pid {
			return &cola[i]
		}
	}
	return nil
}

// func manejarOperacionIO(dispositivo *utilsKernel.IODevice, proceso *kernel.Proceso, tiempoUso int64) {
//     // Simular operación IO (en producción sería llamada HTTP real al módulo)
//     time.Sleep(time.Duration(tiempoUso) * time.Millisecond)
//
//     // Bloquear para actualizar estado del dispositivo
//     dispositivo.Mutex.Lock()
//     defer dispositivo.Mutex.Unlock()
//
//     // Liberar dispositivo
//     dispositivo.Ocupado = false
//     dispositivo.ProcesoEnUso = nil
//
//     // Manejar cola de espera
//     if len(dispositivo.ColaEspera) > 0 {
//         // Tomar siguiente proceso (FIFO)
//         siguienteProceso := dispositivo.ColaEspera[0]
//         dispositivo.ColaEspera = dispositivo.ColaEspera[1:]
//
//         // Asignar dispositivo
//         dispositivo.Ocupado = true
//         dispositivo.ProcesoEnUso = siguienteProceso
//
//         // Notificar al kernel para despertar proceso
//         kernel.NotificarIOLista(siguienteProceso.PID, dispositivo.Nombre)
//     }
// }

