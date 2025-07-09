package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorOrange = "\033[38;5;208m" // naranja aproximado usando color 256
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
	var syscall estructuras.Syscall_Init_Proc

	if err := json.NewDecoder(r.Body).Decode(&syscall); err != nil {
		http.Error(w, "Error al parsear el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	if syscall.ArchivoInstrucciones == "" || syscall.Tamanio <= 0 {
		http.Error(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}

	global.LoggerKernel.Log(ColorGreen+"## ("+strconv.Itoa(syscall.PID)+") - Solicitó syscall: <INIT_PROC>"+ColorReset, log.INFO)
	planificacion.CrearProceso(syscall.Tamanio, syscall.ArchivoInstrucciones)
	//el log ya lo hace crearProceso

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
	}

	global.CPUsConectadas = append(global.CPUsConectadas, &nuevaCpu)
	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", nuevoHandshake.ID, nuevoHandshake.IP, strconv.Itoa(nuevoHandshake.Puerto)), log.DEBUG)

	w.WriteHeader(http.StatusOK)

	pcb := global.NuevoPCB()

	respuesta := global.PCB{
		PID: pcb.PID,
		PC:  pcb.PC,
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

	err := ManejarSolicitudIO(syscall.PIDproceso, syscall.IoSolicitada, syscall.TiempoEstimado)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ManejarSolicitudIO(pid int, nombre string, tiempoUso int) error {
	global.LoggerKernel.Log(ColorBlue+"## ("+strconv.Itoa(pid)+") - Solicitó syscall: <IO>"+ColorReset, log.INFO)

	global.IOListMutex.Lock()
	dispositivos := utilsKernel.ObtenerDispositivoIO(nombre)
	global.IOListMutex.Unlock()

	global.MutexExecuting.Lock() //Muevo
	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaExecuting, pid)
	if proceso == nil {
		global.MutexExecuting.Unlock()//Actualizo esto xq sino nunca sale del lock
		return fmt.Errorf("no se pudo obtener el proceso en EXECUTING (PID %d)", pid)
	}

	
	global.EliminarProcesoDeCola(&global.ColaExecuting, proceso.PID)
	global.MutexExecuting.Unlock()

	if len(dispositivos) == 0 {
		global.LoggerKernel.Log(fmt.Sprintf("Dispositivo IO %s no existe, enviando %d a EXIT", nombre, pid), log.ERROR)
		planificacion.FinalizarProceso(proceso)
		return fmt.Errorf("dispositivo IO %s no existe", nombre)
	}

	procesoEncolado := &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
	global.AgregarABlocked(proceso)

	// buscar un dispositivo libre
	for _, dispositivo := range dispositivos {
		dispositivo.Mutex.Lock()
		if !dispositivo.Ocupado {
			dispositivo.Ocupado = true
			dispositivo.ProcesoEnUso = procesoEncolado
			dispositivo.Mutex.Unlock()

			global.LoggerKernel.Log("## ("+strconv.Itoa(pid)+") - Bloqueado por IO: <"+dispositivo.Nombre+">", log.INFO)
			go utilsKernel.EnviarAIO(dispositivo, pid, tiempoUso)
			return nil
		}
		dispositivo.Mutex.Unlock()
	}

	// Si todos ocupados, encolar en el primero
	primero := dispositivos[0]
	primero.Mutex.Lock()
	primero.ColaEspera = append(primero.ColaEspera, procesoEncolado)
	global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Encolado en %s (Ocupado)", pid, primero.Nombre), log.INFO)
	primero.Mutex.Unlock()

	return nil
}
func FinalizacionIO(w http.ResponseWriter, r *http.Request) {
	// Si NO hay body → desconexión de dispositivo IO
	if r.ContentLength == 0 {
		global.LoggerKernel.Log("[DEBUG] Desconexión recibida", log.DEBUG)

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

		proc := dispositivo.ProcesoEnUso.Proceso

		global.MutexSuspBlocked.Lock()
		global.EliminarProcesoDeCola(&global.ColaSuspBlocked, proc.PID)
		global.MutexSuspBlocked.Unlock()

		global.MutexBlocked.Lock()
		global.EliminarProcesoDeCola(&global.ColaBlocked, proc.PID)
		global.MutexBlocked.Unlock()

		global.IOListMutex.Lock()
		global.IOConectados = utilsKernel.FiltrarIODevice(global.IOConectados, dispositivo)
		global.IOListMutex.Unlock()

		planificacion.FinalizarProceso(proc)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Proceso %d desconectado y marcado como EXIT.\n", proc.PID)
		return
	}

	// Caso normal: finalización de IO
	var mensaje estructuras.FinDeIO
	if err := json.NewDecoder(r.Body).Decode(&mensaje); err != nil {
		http.Error(w, "Error al decodificar mensaje de finalización", http.StatusBadRequest)
		return
	}
	pid := mensaje.PID
	global.LoggerKernel.Log(fmt.Sprintf("[DEBUG] Finalizó IO del PID: %d", pid), log.DEBUG)

	dispositivo := utilsKernel.BuscarDispositivoPorPID(pid)
	if dispositivo == nil {
		http.Error(w, "No se encontró dispositivo para el PID", http.StatusNotFound)
		return
	}

	// Solo manejamos si efectivamente era el proceso en uso
	if dispositivo.ProcesoEnUso != nil && dispositivo.ProcesoEnUso.Proceso.PID == pid {
		proceso := dispositivo.ProcesoEnUso.Proceso

		// 1) Actualizar estado del proceso que terminó IO, fuera del lock del dispositivo
		global.MutexSuspBlocked.Lock()
		enSuspBlocked := global.EliminarProcesoDeCola(&global.ColaSuspBlocked, proceso.PID)
		global.MutexSuspBlocked.Unlock()

		if enSuspBlocked {
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.SUSP_READY)
			global.AgregarASuspReady(proceso)
			select {
			case global.NotifySuspReady <- struct{}{}:
			default: // si ya había señal pendiente, no bloquear
			}
		} else {
			global.MutexBlocked.Lock()
			global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
			global.MutexBlocked.Unlock()

			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)
			global.AgregarAReady(proceso)
			global.LoggerKernel.Log("AGREGAR A READY A", log.DEBUG)
		}

		// 2) Ahora, con el dispositivo liberado, asignar IO al siguiente en la cola
		dispositivo.Mutex.Lock()
		if len(dispositivo.ColaEspera) > 0 {
			nuevo := dispositivo.ColaEspera[0]
			//global.LoggerKernel.Log(fmt.Sprintf("## (%d) - Proceso en espera: %d", pid, nuevo.Proceso.PID), log.INFO)

			// Remover de la cola de espera y asignar
			dispositivo.ColaEspera = utilsKernel.FiltrarColaIO(dispositivo.ColaEspera, nuevo)
			dispositivo.ProcesoEnUso = nuevo
			dispositivo.Mutex.Unlock()
			utilsKernel.EnviarAIO(dispositivo, nuevo.Proceso.PID, nuevo.TiempoUso)
		} else {
			// Si no hay más en espera, el dispositivo queda libre
			dispositivo.ProcesoEnUso = nil
			dispositivo.Ocupado = false
			dispositivo.Mutex.Unlock()
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "No se encontró proceso en uso para este PID", http.StatusNotFound)
}

func EXIT(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	PID, _ := strconv.Atoi(pidStr)

	global.LoggerKernel.Log(ColorRed+"## ("+strconv.Itoa(PID)+") - Solicitó syscall: <EXIT>"+ColorReset, log.INFO)

	w.WriteHeader(http.StatusOK)
}

func DUMP_MEMORY(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	pid, _ := strconv.Atoi(pidStr)

	global.LoggerKernel.Log(ColorOrange+"## ("+strconv.Itoa(pid)+") - Solicitó syscall: <DUMP_MEMORY>"+ColorReset, log.INFO)
	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaBlocked, pid)
	if proceso == nil {
		global.LoggerKernel.Log(fmt.Sprintf("ERROR: No se encontró el proceso con PID %d en ColaBlocked", pid), log.ERROR)
		http.Error(w, "Proceso no encontrado", http.StatusNotFound)
		return
	}
	// Solicitar dump
	err := utilsKernel.SolicitarDumpAMemoria(pid)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", pid, err.Error()), log.ERROR)

		global.MutexBlocked.Lock()
		global.EliminarProcesoDeCola(&global.ColaBlocked, pid)
		global.MutexBlocked.Unlock()

		planificacion.FinalizarProceso(proceso)

		http.Error(w, "Fallo en Dump, proceso finalizado", http.StatusInternalServerError)
		return
	}

	// Volver a READY
	global.MutexBlocked.Lock()
	global.EliminarProcesoDeCola(&global.ColaBlocked, pid)
	global.MutexBlocked.Unlock()

	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)
	global.AgregarAReady(proceso)
	global.LoggerKernel.Log("AGREGAR A READY B", log.DEBUG)

	w.WriteHeader(http.StatusOK)
}

func DevolucionCPUHandler(w http.ResponseWriter, r *http.Request) {
	var devolucion estructuras.RespuestaCPU
	err := json.NewDecoder(r.Body).Decode(&devolucion)
	if err != nil {
		http.Error(w, "Error al decodificar devolución", http.StatusBadRequest)
		return
	}

	go planificacion.ManejarDevolucionDeCPU(devolucion)
	w.WriteHeader(http.StatusOK)

	global.LoggerKernel.Log(fmt.Sprintf("Llega PID %d y PC %d", devolucion.PID, devolucion.PC), log.DEBUG)
}
