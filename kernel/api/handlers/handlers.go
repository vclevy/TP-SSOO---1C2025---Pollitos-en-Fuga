package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
	"time"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO
type IODevice = global.IODevice
type PCB = planificacion.PCB
type SyscallIO = estructuras.Syscall_IO
type FinDeIO = estructuras.FinDeIO
type Syscall_Init_Proc = estructuras.Syscall_Init_Proc

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
		//global.LoggerKernel.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
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
	var syscall estructuras.Syscall_Init_Proc

	if err := json.NewDecoder(r.Body).Decode(&syscall); err != nil {
		http.Error(w, "Error al parsear el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	if syscall.ArchivoInstrucciones == "" || syscall.Tamanio < 0 {
		http.Error(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}

	global.LoggerKernel.Log("## ("+strconv.Itoa(syscall.PID)+") - Solicitó syscall: <INIT_PROC>", log.INFO)
	planificacion.CrearProceso(syscall.Tamanio, syscall.ArchivoInstrucciones)
}

func HandshakeConCPU(w http.ResponseWriter, r *http.Request) {
	var nuevoHandshake estructuras.HandshakeConCPU
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&nuevoHandshake); err != nil {
		http.Error(w, "Body inválido", http.StatusBadRequest)
		return
	}

	nuevaCpu := global.CPU{
		ID:                nuevoHandshake.ID,
		IP:                nuevoHandshake.IP,
		Puerto:            nuevoHandshake.Puerto,
		ProcesoEjecutando: nil,
		CanalProceso:      make(chan *global.Proceso),
	}

	global.MutexCPUs.Lock()
	global.CPUsConectadas = append(global.CPUsConectadas, &nuevaCpu)
	global.MutexCPUs.Unlock()

	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", nuevoHandshake.ID, nuevoHandshake.IP, strconv.Itoa(nuevoHandshake.Puerto)), log.DEBUG)

	w.WriteHeader(http.StatusOK)

	pcb := global.NuevoPCB()

	respuesta := global.PCB{
		PID: pcb.PID,
		PC:  pcb.PC,
	}

	go LoopCPU(&nuevaCpu)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

func LoopCPU(cpu *global.CPU) {
	for {
		proceso := <-cpu.CanalProceso
		if proceso.PCB.UltimoEstado != planificacion.EXEC {
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXEC)
		}
		global.AgregarAExecuting(proceso)
		proceso.InstanteInicio = time.Now()
		//global.LoggerKernel.Log(fmt.Sprintf("CPU %s: Recibió proceso PID %d", cpu.ID, proceso.PCB.PID), log.DEBUG)

		err := utilsKernel.EnviarADispatch(cpu, proceso.PCB.PID, proceso.PCB.PC)
		if err != nil {
			global.LoggerKernel.Log(fmt.Sprintf("Error enviando proceso PID %d a Dispatch: %s", proceso.PCB.PID, err.Error()), log.ERROR)
			continue
		}
	}
}

func FinalizacionIO(w http.ResponseWriter, r *http.Request) {
	// sin body: desconexión del io
	if r.ContentLength == 0 {
		global.LoggerKernel.Log("Desconexión recibida", log.DEBUG)

		ip := r.Header.Get("X-IO-IP")
		puertoStr := r.Header.Get("X-IO-Puerto")
		puerto, err := strconv.Atoi(puertoStr)
		if err != nil {
			http.Error(w, "Puerto inválido en header", http.StatusBadRequest)
			return
		}

		dispositivo, err := utilsKernel.BuscarDispositivo(ip, puerto)
		if err != nil {
			http.Error(w, "Dispositivo no encontrado en desconexión", http.StatusNotFound)
			return
		}

		if dispositivo.ProcesoEnUso == nil {
			global.LoggerKernel.Log(fmt.Sprintf("IO %s:%d se desconectó sin proceso en uso", ip, puerto), log.DEBUG)
		} else {
			proc := dispositivo.ProcesoEnUso.Proceso

			global.MutexSuspBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaSuspBlocked, proc.PID)
			global.MutexSuspBlocked.Unlock()

			global.MutexBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaBlocked, proc.PID)
			global.MutexBlocked.Unlock()

			planificacion.FinalizarProceso(proc)
		}

		global.IOListMutex.Lock()
		global.IOConectados = utilsKernel.FiltrarIODevice(global.IOConectados, dispositivo)
		global.IOListMutex.Unlock()

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "IO %s:%d desconectado.\n", ip, puerto)
		return
	}

	// con body: se termina normal
	var mensaje estructuras.FinDeIO
	if err := json.NewDecoder(r.Body).Decode(&mensaje); err != nil {
		http.Error(w, "Error al decodificar mensaje de finalización", http.StatusBadRequest)
		return
	}
	pid := mensaje.PID
	//global.LoggerKernel.Log(fmt.Sprintf("Finalizó IO del PID: %d", pid), log.DEBUG)

	dispositivo := utilsKernel.BuscarDispositivoPorPID(pid)
	if dispositivo == nil {
		http.Error(w, "No se encontró dispositivo para el PID", http.StatusNotFound)
		return
	}

	if dispositivo.ProcesoEnUso != nil && dispositivo.ProcesoEnUso.Proceso.PID == pid {
		proceso := dispositivo.ProcesoEnUso.Proceso

		global.MutexSuspBlocked.Lock()
		enSuspBlocked := global.EliminarProcesoDeCola(&global.ColaSuspBlocked, proceso.PID)
		global.MutexSuspBlocked.Unlock()

		if enSuspBlocked {
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.SUSP_READY)
			global.AgregarASuspReady(proceso)
			select {
			case global.NotifySuspReady <- struct{}{}:
			default:
			}
		} else {
			global.MutexBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
			global.MutexBlocked.Unlock() //
			global.LoggerKernel.Log(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", proceso.PID), log.INFO)
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)
			planificacion.AgregarAReady(proceso)

		}

		dispositivo.Mutex.Lock()
		if len(dispositivo.ColaEspera) > 0 {
			nuevo := dispositivo.ColaEspera[0]
			dispositivo.ColaEspera = utilsKernel.FiltrarColaIO(dispositivo.ColaEspera, nuevo)
			dispositivo.ProcesoEnUso = nuevo
			dispositivo.Mutex.Unlock()
			utilsKernel.EnviarAIO(dispositivo, nuevo.Proceso.PID, nuevo.TiempoUso)
		} else {
			dispositivo.ProcesoEnUso = nil
			dispositivo.Ocupado = false
			dispositivo.Mutex.Unlock()
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "No se encontró proceso en uso para este PID", http.StatusNotFound)
}

func DevolucionCPUHandler(w http.ResponseWriter, r *http.Request) {
	var devolucion estructuras.RespuestaCPU
	err := json.NewDecoder(r.Body).Decode(&devolucion)
	if err != nil {
		http.Error(w, "Error al decodificar devolución", http.StatusBadRequest)
		return
	}

	planificacion.ManejarDevolucionDeCPU(devolucion)
	w.WriteHeader(http.StatusOK)

	//global.LoggerKernel.Log(fmt.Sprintf("Llega PID %d y PC %d", devolucion.PID, devolucion.PC), log.DEBUG)
}
