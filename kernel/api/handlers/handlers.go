package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion "github.com/sisoputnfrba/tp-golang/kernel/planificacion"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
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

	global.LoggerKernel.Log(ColorGreen + "## ("+strconv.Itoa(syscall.PID)+") - Solicitó syscall: <INIT_PROC>" + ColorReset, log.INFO)
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

	nombre := strings.ToLower(syscall.IoSolicitada)
	tiempoUso := syscall.TiempoEstimado
	pid := syscall.PIDproceso

	global.LoggerKernel.Log(ColorBlue + "## ("+strconv.Itoa(pid)+") - Solicitó syscall: <IO>" + ColorReset, log.INFO)

	global.IOListMutex.Lock()
	dispositivos := utilsKernel.ObtenerDispositivoIO(nombre)
	global.IOListMutex.Unlock()

	
	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaBlocked, pid)
	if proceso == nil {
		http.Error(w, "No se pudo obtener el proceso en BLOCKED", http.StatusInternalServerError)
		return
	}

	if len(dispositivos) == 0 {
		global.LoggerKernel.Log(fmt.Sprintf("Dispositivo IO %s no existe, enviando %d a exit", nombre, pid), log.ERROR)
		global.MutexBlocked.Lock()
		global.EliminarProcesoDeCola(&global.ColaBlocked,proceso.PID)
		global.MutexBlocked.Unlock()
		planificacion.FinalizarProceso(proceso)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	
	procesoEncolado := &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

	//buscar un dispositivo libre
	for _, dispositivo := range dispositivos {
		dispositivo.Mutex.Lock()
		if !dispositivo.Ocupado {
			dispositivo.Ocupado = true
			dispositivo.ProcesoEnUso = procesoEncolado
			dispositivo.Mutex.Unlock()

			global.LoggerKernel.Log("## ("+strconv.Itoa(pid)+") - Bloqueado por IO: <"+dispositivo.Nombre+">", log.INFO)
			go utilsKernel.EnviarAIO(dispositivo, pid, tiempoUso)

			w.WriteHeader(http.StatusOK)
			return
		}
		dispositivo.Mutex.Unlock()
	}

	// Si nadie esta libre, encolar en el primero
	primero := dispositivos[0]
	primero.Mutex.Lock()
	primero.ColaEspera = append(primero.ColaEspera, procesoEncolado)
	primero.Mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func FinalizacionIO(w http.ResponseWriter, r *http.Request) {

	// Si NO hay body desconexión
	if r.ContentLength == 0 {
		fmt.Println("Desconexión recibida")

		r.Header.Get("X-IO-Nombre")
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
	global.LoggerKernel.Log(fmt.Sprintf("Finalizó IO del PID: %d\n", pid), log.DEBUG)

	dispositivo := utilsKernel.BuscarDispositivoPorPID(pid)
	if dispositivo == nil {
		http.Error(w, "No se encontró dispositivo para el PID", http.StatusNotFound)
		return
	}

	if dispositivo.ProcesoEnUso != nil && dispositivo.ProcesoEnUso.Proceso.PID == pid {
		proceso := dispositivo.ProcesoEnUso.Proceso
		dispositivo.ProcesoEnUso = nil

		// Primero intentar sacar de SUSP_BLOCKED
		global.MutexSuspBlocked.Lock()
		enSuspBlocked := global.EliminarProcesoDeCola(&global.ColaSuspBlocked, proceso.PID)
		global.MutexSuspBlocked.Unlock()

		if enSuspBlocked {
			global.AgregarASuspReady(proceso)
			planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.SUSP_READY)

			global.LoggerKernel.Log(fmt.Sprintf("## (%d) finalizó IO y pasa a SUSP_READY", proceso.PID), log.INFO)

			select {
			case global.NotifySuspReady <- struct{}{}:
				global.LoggerKernel.Log("Notificando a largo plazo por SUSP_READY", log.DEBUG)
			default:
			}
			return
		}

		// Si no estaba en SUSP_BLOCKED, asumimos que estaba en BLOCKED
		global.MutexBlocked.Lock()
		global.EliminarProcesoDeCola(&global.ColaBlocked, proceso.PID)
		global.MutexBlocked.Unlock()

		global.AgregarAReady(proceso)
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)

		global.LoggerKernel.Log(fmt.Sprintf("## (%d) finalizó IO y pasa a READY", proceso.PID), log.INFO)

		select {
		case global.NotifyReady <- struct{}{}:
		default:
		}
		// Ver si hay proceso esperando en el dispositivo
		dispositivo.Mutex.Lock()
		if len(dispositivo.ColaEspera) > 0 {
			nuevo := dispositivo.ColaEspera[0]
			dispositivo.ColaEspera = utilsKernel.FiltrarColaIO(dispositivo.ColaEspera, nuevo)
			dispositivo.ProcesoEnUso = nuevo
			dispositivo.Mutex.Unlock()

			utilsKernel.EnviarAIO(dispositivo, nuevo.Proceso.PID, nuevo.TiempoUso)
		} else {
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

	global.LoggerKernel.Log(ColorRed + "## ("+strconv.Itoa(PID)+") - Solicitó syscall: <EXIT>" + ColorReset, log.INFO)

	w.WriteHeader(http.StatusOK)
}

func DUMP_MEMORY(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	pid, _ := strconv.Atoi(pidStr)

	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaExecuting, pid)

	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
	global.AgregarABlocked(proceso)
	global.LoggerKernel.Log(ColorOrange + "## ("+strconv.Itoa(pid)+") - Solicitó syscall: <DUMP_MEMORY>" + ColorReset, log.INFO)

	err := utilsKernel.SolicitarDumpAMemoria(pid)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", pid, err.Error()), log.ERROR)

		global.EliminarProcesoDeCola(&global.ColaBlocked, pid)
		planificacion.FinalizarProceso(proceso)

		http.Error(w, "Fallo en Dump, proceso finalizado", http.StatusInternalServerError)
		return
	}

	global.EliminarProcesoDeCola(&global.ColaBlocked, pid)
	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.READY)
	global.AgregarAReady(proceso)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Dump exitoso para PID %d", pid)
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
}
