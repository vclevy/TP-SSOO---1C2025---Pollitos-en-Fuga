package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	//"time"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO
type IODevice = global.IODevice
type PCB = planificacion.PCB
type SyscallIO = estructuras.Syscall_IO
type FinDeIO = estructuras.FinDeIO
type Syscall_Init_Proc = estructuras.Syscall_Init_Proc

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
    var syscall estructuras.Syscall_Init_Proc

	if err := json.NewDecoder(r.Body).Decode(&syscall); err != nil {
		http.Error(w, "Error al parsear el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	if syscall.ArchivoInstrucciones == "" || syscall.Tamanio <= 0 {
		http.Error(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}

	procesoCreado := planificacion.CrearProceso(syscall.Tamanio, syscall.ArchivoInstrucciones)

    global.LoggerKernel.Log(fmt.Sprintf("Proceso creado: %+v", procesoCreado), log.DEBUG)
	global.MutexNew.Lock()
	global.ColaNew = append(global.ColaNew, procesoCreado)
	global.MutexNew.Unlock()
}


func HandshakeConCPU(w http.ResponseWriter, r *http.Request) { //Solo conexion inicial
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
		ID:               nuevoHandshake.ID,
		IP:               nuevoHandshake.IP,
		Puerto:           nuevoHandshake.Puerto,
		ProcesoEjecutando: nil,
	}

	global.CPUsConectadas = append(global.CPUsConectadas, &nuevaCpu)
	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", nuevoHandshake.ID, nuevoHandshake.IP, strconv.Itoa(nuevoHandshake.Puerto)), log.DEBUG)

	w.WriteHeader(http.StatusOK)
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

	proceso := BuscarProcesoPorPID(global.ColaExecuting, pid)
	if proceso == nil {
		http.Error(w, "No se pudo obtener el proceso actual", http.StatusInternalServerError)
		return
	}
	if len(dispositivos) == 0 {
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXIT)
		global.MutexExit.Lock()
		global.ColaExit = append(global.ColaExit, proceso)
		global.MutexExit.Unlock()
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Dispositivo %s no existe", nombre)
		return
	}

	for _, dispositivo := range dispositivos {
		dispositivo.Mutex.Lock()
		if !dispositivo.Ocupado {
			dispositivo.Ocupado = true
			dispositivo.ProcesoEnUso = &global.ProcesoIO{
				Proceso:   proceso,
				TiempoUso: tiempoUso,
			}
			dispositivo.Mutex.Unlock()

			go utilsKernel.EnviarAIO(dispositivo, proceso.PCB.PID, tiempoUso)
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
			return
		}
		dispositivo.Mutex.Unlock()
	}

	// Si todos están ocupados, se encola en el primero
	procesoEncolado := &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

	primero := dispositivos[0]
	primero.Mutex.Lock()
	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
	primero.ColaEspera = append(primero.ColaEspera, procesoEncolado)
	primero.Mutex.Unlock()
}


func BuscarProcesoPorPID(cola []*global.Proceso, pid int) (*global.Proceso) {
	for i := range cola {
		if cola[i].PCB.PID == pid {
			return cola[i]
		}
	}
	return nil
}


func FinalizacionIO(w http.ResponseWriter, r *http.Request) {
	host, portStr, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Error al parsear dirección remota", http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Puerto inválido", http.StatusBadRequest)
		return
	}

	dispositivo, err := utilsKernel.BuscarDispositivo(host, port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Verificamos si hay body. Si NO hay, es desconexión
	if r.ContentLength == 0 {
		proc := dispositivo.ProcesoEnUso.Proceso
		global.MutexBlocked.Lock()
		global.ColaBlocked = utilsKernel.FiltrarCola(global.ColaBlocked, proc)
		global.MutexBlocked.Unlock()
		global.IOListMutex.Lock()
		global.IOConectados = utilsKernel.FiltrarIODevice(global.IOConectados, dispositivo)
		global.IOListMutex.Unlock()
		planificacion.FinalizarProceso(proc)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Proceso %d desconectado y marcado como EXIT.\n", proc.PID)
		return
	}

	// Caso normal: Fin de IO
	if dispositivo.ProcesoEnUso != nil {
		fmt.Fprintf(w, "Proceso %d completó E/S. Verificando cola...\n", dispositivo.ProcesoEnUso.Proceso.PID)

		// Liberamos proceso actual
		dispositivo.ProcesoEnUso = nil

		if len(dispositivo.ColaEspera) > 0 {
			dispositivo.Mutex.Lock()
			nuevo := dispositivo.ColaEspera[0]
			dispositivo.ColaEspera = utilsKernel.FiltrarColaIO(dispositivo.ColaEspera, nuevo)
			dispositivo.Mutex.Unlock()
			dispositivo.ProcesoEnUso = nuevo

			utilsKernel.EnviarAIO(dispositivo, nuevo.Proceso.PID, nuevo.TiempoUso)
		} else {
			dispositivo.Ocupado = false
		}
	}

	w.WriteHeader(http.StatusOK)
}

func EXIT(w http.ResponseWriter, r *http.Request){
	pidStr := r.URL.Query().Get("pid")

	PID,_ := strconv.Atoi(pidStr)
	proceso := BuscarProcesoPorPID(global.ColaExecuting, PID)

	if proceso == nil {
		http.Error(w, "Proceso no encontrado", http.StatusNotFound)
		return
	}
	planificacion.FinalizarProceso(proceso)
	//* En finalizarProceso se actualiza el pcb y se mueve a la cola correspondiente
}


func DUMP_MEMORY(w http.ResponseWriter, r *http.Request){
	pidStr := r.URL.Query().Get("pid")
	pid,_ := strconv.Atoi(pidStr)

	proceso := BuscarProcesoPorPID(global.ColaExecuting,pid)
	
	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)

	global.MutexBlocked.Lock()
	global.ColaBlocked = append(global.ColaBlocked, proceso)
	global.MutexBlocked.Unlock()

	err := utilsKernel.SolicitarDumpAMemoria(pid)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", pid, err.Error()), log.ERROR)
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXIT)
		global.MutexExit.Lock()
		global.ColaExit = append(global.ColaExit, proceso)
		global.MutexExit.Unlock()
		planificacion.FinalizarProceso(proceso)
		http.Error(w, "Fallo en Dump, proceso finalizado", http.StatusInternalServerError)
		return
	}

	// Si todo va bien, pasa a READY
	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)
	global.MutexReady.Lock()
	global.ColaReady = append(global.ColaReady, proceso)
	global.MutexReady.Unlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Dump exitoso para PID %d", pid)
}

//APIs para conexion con cada instancia de CPU

func FinalizarProceso(w http.ResponseWriter, r *http.Request){
	//TODO: Recibe notificación de que un proceso terminó
}

func DevolverPCB(w http.ResponseWriter, r *http.Request){
	//TODO: si querés interrumpir y recuperar el estado del proceso 
}