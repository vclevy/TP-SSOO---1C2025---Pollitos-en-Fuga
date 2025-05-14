package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
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
		ID:               nuevoHandshake.ID,
		IP:               nuevoHandshake.IP,
		Puerto:           nuevoHandshake.Puerto,
		ProcesoEjecutando: nil,
	}

	global.CPUsConectadas = append(global.CPUsConectadas, &nuevaCpu)
	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", nuevoHandshake.ID, nuevoHandshake.IP, strconv.Itoa(nuevoHandshake.Puerto)), log.DEBUG)

	w.WriteHeader(http.StatusOK)

	pcb := global.NuevoPCB()

	respuesta := map[string]int{
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

func manejarIOOcupado(io *global.IODevice, proceso *global.Proceso, tiempoUso int, w http.ResponseWriter) {
	// Agregar a cola de espera
	io.ColaEspera = append(io.ColaEspera, &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	})

	// Cambiar estado según si está en memoria o swap
	if proceso.PCB.UltimoEstado == planificacion.SUSP_READY || proceso.PCB.UltimoEstado == planificacion.SUSP_BLOCKED {
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.SUSP_BLOCKED)
	} else {
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.BLOCKED)
		global.ColaBlocked = append(global.ColaBlocked, proceso)
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Proceso %d en cola para %s", proceso.PID, io.Nombre)
}


func manejarIOLibre(io *global.IODevice, proceso *global.Proceso, tiempoUso int, w http.ResponseWriter) {
	// Asignar dispositivo
	io.Ocupado = true
	io.ProcesoEnUso = &global.ProcesoIO{
		Proceso:   proceso,
		TiempoUso: tiempoUso,
	}

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
			proxProceso := siguiente.Proceso

			global.MutexSuspBlocked.Lock()
			// Verificar si el proceso está suspendido
			for i, p := range global.ColaSuspBlocked {
				if p.PID == proxProceso.PID {
					planificacion.ActualizarEstadoPCB(&p.PCB, planificacion.SUSP_READY)
					global.MutexSuspReady.Lock()
					global.ColaSuspReady = append(global.ColaSuspReady, p)
					global.MutexSuspReady.Unlock()
					global.ColaSuspBlocked = append(global.ColaSuspBlocked[:i], global.ColaSuspBlocked[i+1:]...)
					return
				}
			}
			global.MutexSuspBlocked.Unlock()

			// Si no estaba suspendido, mover a READY
			planificacion.ActualizarEstadoPCB(&proxProceso.PCB, planificacion.READY)
			global.MutexReady.Lock()
			global.ColaReady = append(global.ColaReady, proxProceso)
			global.MutexReady.Unlock()


			global.MutexBlocked.Lock()
			// Remover de BLOCKED si estaba allí
			for i, p := range global.ColaBlocked {
				if p.PID == proxProceso.PID {
					global.ColaBlocked = append(global.ColaBlocked[:i], global.ColaBlocked[i+1:]...)
					break
				}
			}
			global.MutexBlocked.Unlock()
			
		}
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Proceso %d accediendo a %s por %d ms", proceso.PID, io.Nombre, tiempoUso)
}


func BuscarProcesoPorPID(cola []*global.Proceso, pid int) (*global.Proceso) {
	for i := range cola {
		if cola[i].PCB.PID == pid {
			return cola[i]
		}
	}
	return nil
}


//! Falta esto creo @valenchu: Al momento que se conecte una nueva IO o se reciba el desbloqueo por medio de una de ellas, se deberá verificar si hay proceso encolados para dicha IO y enviarlo a la misma. 

func FinalizacionIO(w http.ResponseWriter, r *http.Request){

	// Extraer IP y puerto del remitente
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

    //? Buscar dispositivo por puerto e ip
    dispositivo, err := utilsKernel.BuscarDispositivo(host, port)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

	var finDeIo FinDeIO
	if err := json.NewDecoder(r.Body).Decode(&finDeIo); err != nil {
		http.Error(w, "Error al parsear la syscall", http.StatusBadRequest)
		return
	}

	tipo := finDeIo.Tipo
	

	switch tipo {
    case "FIN_IO":
        // Lógica para FIN_IO
        // 1. Verificar si hay más procesos en cola de I/O
		fmt.Fprintf(w, "Proceso %d completó E/S. Verificando cola...", dispositivo.ProcesoEnUso.Proceso.PID)
		//! @valenchu Aca va la parte del planificador de mediano plazo creom tiene que pasar a susp ready?
		dispositivo.ProcesoEnUso = nil //saco el proceso actual
		if dispositivo.ColaEspera != nil{

			nuevoProcesoEnIO := dispositivo.ColaEspera[0]
			dispositivo.ColaEspera = utilsKernel.FiltrarColaIO(dispositivo.ColaEspera,nuevoProcesoEnIO)
			dispositivo.ProcesoEnUso = nuevoProcesoEnIO
			pidNuevo := dispositivo.ProcesoEnUso.Proceso.PID
			utilsKernel.EnviarAIO(dispositivo, pidNuevo, nuevoProcesoEnIO.TiempoUso)
			//? que se hace si hay mas procesos deberian ejecutarse? esta bien esto q hice?
			
		} else{
			dispositivo.Ocupado = false
		}
    
        w.WriteHeader(http.StatusOK)

    case "DESCONEXION_IO":
        // Lógica para DESCONEXION_IO
        // 1. Cambiar estado del proceso a EXIT
		global.ColaBlocked = utilsKernel.FiltrarCola(global.ColaBlocked, dispositivo.ProcesoEnUso.Proceso)
		global.ColaExit = append(global.ColaExit, dispositivo.ProcesoEnUso.Proceso)
		planificacion.ActualizarEstadoPCB(&dispositivo.ProcesoEnUso.Proceso.PCB, planificacion.EXIT)
		//! Chequeame esto @valenchu

        // Ejemplo:
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Proceso %d desconectado de E/S. Marcado como EXIT.", dispositivo.ProcesoEnUso.Proceso.PID)

    default:
        http.Error(w, "Tipo de operación no válido", http.StatusBadRequest)
    }
	
}

func EXIT(w http.ResponseWriter, r *http.Request){
	pidStr := r.URL.Query().Get("pid")

	PID,_ := strconv.Atoi(pidStr)
	proceso := BuscarProcesoPorPID(global.ColaExecuting, PID)

	if proceso == nil {
		http.Error(w, "Proceso no encontrado", http.StatusNotFound)
		return
	}

	planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXIT)
	global.MutexExit.Lock()
	global.ColaExit = append(global.ColaExit, proceso)
	global.MutexExit.Unlock()	
}

func DUMP_MEMORY(w http.ResponseWriter, r *http.Request){
	pidStr := r.URL.Query().Get("pid")
	pid,_ := strconv.Atoi(pidStr)

	proceso := BuscarProcesoPorPID(global.ColaExecuting,pid)

	err := utilsKernel.SolicitarDumpAMemoria(pid)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error en dump de memoria para PID %d: %s", pid, err.Error()), log.ERROR)
		planificacion.ActualizarEstadoPCB(&proceso.PCB, planificacion.EXIT)
		global.MutexExit.Lock()
		global.ColaExit = append(global.ColaExit, proceso)
		global.MutexExit.Unlock()
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