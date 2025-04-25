package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	planificacion"github.com/sisoputnfrba/tp-golang/kernel/planificacion"

	"github.com/sisoputnfrba/tp-golang/utils/logger"
	utils "github.com/sisoputnfrba/tp-golang/utils/paquetes"
	conexionConIO "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
}

type Respuesta struct {
	Status        string `json:"status"`
	Detalle       string `json:"detalle"`
	PID           int    `json:"pid"`
	TiempoEstimado int   `json:"tiempo_estimado"`
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

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerKernel.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}

	global.LoggerKernel.Log("Kernel recibió paquete desde IO - Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" | Código: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	// Simulación de asignación de PID y tiempo
	fmt.Println("Ingrese el PID: ")
	str_pid := utils.LeerStringDeConsola()
	fmt.Println("Ingrese el tiempo estimado: ")
	str_tiempoEstimado := utils.LeerStringDeConsola()
	
	pid, _:= strconv.Atoi(str_pid)
	tiempoEstimado, _ :=  strconv.Atoi(str_tiempoEstimado); // no se si conviene hacer este cambio o que leerstring lea int

	respuesta := Respuesta{
		Status:        "OK",
		Detalle:       "Paquete procesado correctamente",
		PID:           pid,
		TiempoEstimado:	tiempoEstimado}

	global.LoggerKernel.Log("Kernel responde a IO: PID="+strconv.Itoa(pid)+", Tiempo="+strconv.Itoa(tiempoEstimado)+"ms", log.DEBUG)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}


type PCB = planificacion.PCB

// Vamos a necesitar aca una api con w*responseWritter y eso para el handler que contiene la func crear proceso


func INIT_PROC(w http.ResponseWriter, r *http.Request){
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

	id:= r.URL.Query().Get("id")
	ip:= r.URL.Query().Get("ip")
	puerto := r.URL.Query().Get("puerto")

	global.LoggerKernel.Log(fmt.Sprintf("Handshake recibido de CPU %s en %s:%s", id, ip, puerto), log.DEBUG)
	
	w.WriteHeader(http.StatusOK)

	pcb := global.NuevoPCB()

	// Podrías guardarlo si lo necesitás más adelante
	// procesos[nuevoPID] = pcb

	// Respondemos a la CPU con los datos del PCB
	respuesta := map[string]interface{}{
		"pid": pcb.PID,
		"pc":  pcb.PC,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

type IOSyscallRequest struct {
	PID           int    `json:"pid"`
	NombreDispositivo string `json:"dispositivo"`
	Duracion      int    `json:"duracion"` // en milisegundos
}

var mutex sync.Mutex

func SyscallIO(w http.ResponseWriter, r *http.Request) {
	var req IOSyscallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if !conexionConIO.DispositivoExiste(req.NombreDispositivo) {
		global.LoggerKernel.Log(fmt.Sprintf("PID %d solicitó IO %s que no existe. Finalizando.", req.PID, req.NombreDispositivo), logger.ERROR)
		conexionConIO.FinalizarProcesoPorSyscall(req.PID) // EXIT directo
		w.WriteHeader(http.StatusOK)
		return
	}

	global.LoggerKernel.Log(fmt.Sprintf("PID %d solicita IO %s por %d ms", req.PID, req.NombreDispositivo, req.Duracion), logger.INFO)

	conexionConIO.BloquearProceso(req.PID, req.NombreDispositivo)
	conexionConIO.EncolarEnIO(req.NombreDispositivo, req.PID, req.Duracion)

	if conexionConIO.DispositivoLibre(req.NombreDispositivo) {
		conexionConIO.EnviarAIO(req.NombreDispositivo)
	}

	w.WriteHeader(http.StatusOK)
}

type DispositivoIO = global.DispositivoIO

func HandleIORegister(w http.ResponseWriter, r *http.Request) {  //*Chequear
	var nuevoIO DispositivoIO
	if err := json.NewDecoder(r.Body).Decode(&nuevoIO); err != nil {
		http.Error(w, "Error al decodificar", http.StatusBadRequest)
		return
	}

	global.DispositivosIO[nuevoIO.Nombre] = nuevoIO

	global.LoggerKernel.Log(fmt.Sprintf("Dispositivo IO registrado: %s (%s:%d)",
		nuevoIO.Nombre, nuevoIO.IP, nuevoIO.Puerto), log.DEBUG)

	w.WriteHeader(http.StatusOK)
}
