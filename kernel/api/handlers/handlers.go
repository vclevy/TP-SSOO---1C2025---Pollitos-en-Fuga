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
	"github.com/sisoputnfrba/tp-golang/utils/logger"
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

	procesoCreado := planificacion.CrearProceso(syscall.Tamanio, syscall.ArchivoInstrucciones)

    global.LoggerKernel.Log(fmt.Sprintf("Proceso creado: %+v", procesoCreado), log.DEBUG)
	//el log ya lo hace crearProceso
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

	nombre := syscall.IoSolicitada
	tiempoUso := syscall.TiempoEstimado
	pid := syscall.PIDproceso

	global.IOListMutex.RLock()
	dispositivos := utilsKernel.ObtenerDispositivoIO(nombre)
	global.IOListMutex.RUnlock()

	if len(dispositivos) == 0 {
		global.LoggerKernel.Log(fmt.Sprintf("Dispositivo IO %s no existe", nombre), log.ERROR)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaBlocked, pid)
	if proceso == nil {
		http.Error(w, "No se pudo obtener el proceso en BLOCKED", http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusOK)
			return
		}
		dispositivo.Mutex.Unlock()
	}

	// Todos ocupados, encolarlo
	procesoEncolado := &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

	primero := dispositivos[0]
	primero.Mutex.Lock()
	primero.ColaEspera = append(primero.ColaEspera, procesoEncolado)
	primero.Mutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func EXIT(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	PID, _ := strconv.Atoi(pidStr)

	global.LoggerKernel.Log(fmt.Sprintf("Recibida syscall EXIT para PID %d", PID), log.DEBUG)

	w.WriteHeader(http.StatusOK) // !! @Delfi : ya no se finaliza el proceso acá, solo responde OK a la cpu y el proceso lo terminamos en planificacion, con manejarDevolucionDeCPU
}

func DUMP_MEMORY(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	pid, _ := strconv.Atoi(pidStr)

	proceso := utilsKernel.BuscarProcesoPorPID(global.ColaExecuting, pid)

	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
	global.AgregarABlocked(proceso)

	// 2. Enviar solicitud a memoria
	err := utilsKernel.SolicitarDumpAMemoria(pid)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", pid, err.Error()), log.ERROR)

		// Eliminar de BLOCKED (ya que no saldrá por vía normal)
		global.EliminarProcesoDeCola(&global.ColaBlocked, pid)
		planificacion.FinalizarProceso(proceso)

		http.Error(w, "Fallo en Dump, proceso finalizado", http.StatusInternalServerError)
		return
	}

	// 3. Si el dump fue exitoso, desbloquear
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

	go planificacion.ManejarDevolucionDeCPU(devolucion.PID, devolucion.PC, devolucion.Motivo, devolucion.RafagaReal)
	w.WriteHeader(http.StatusOK)
}
