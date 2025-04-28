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

	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	//conexionConIO "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO

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

	var paquete PaqueteHandshakeIO
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerKernel.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}

	global.LoggerKernel.Log("Kernel recibió paquete desde IO - Nombre: "+ paquete.NombreIO+" | IP IO: "+paquete.IPIO+" | Puerto Io: "+ strconv.Itoa(paquete.PuertoIO), log.DEBUG)

	// // Simulación de asignación de PID y tiempo
	// respuesta := Respuesta{
	// 	Status:        "OK",
	// 	Detalle:       "Paquete procesado correctamente",
	// 	PID:           pid,
	// 	TiempoEstimado:	tiempoEstimado}

	// global.LoggerKernel.Log("Kernel responde a IO: PID="+strconv.Itoa(pid)+", Tiempo="+strconv.Itoa(tiempoEstimado)+"ms", log.DEBUG)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(respuesta)
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

type MensajeIO struct {
	NombreIO string `json:"nombre_io"`
	Evento   string `json:"evento"`   // "registro", "fin", "desconexion"
	PID      int    `json:"pid"`      // Opcional, solo si es fin
	IP       string `json:"ip"`       // Solo para registro
	Puerto   int    `json:"puerto"`   // Solo para registro
}

func ManejarMensajeIO(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error leyendo body", http.StatusBadRequest)
		return
	}

	var mensaje MensajeIO
	err = json.Unmarshal(body, &mensaje)
	if err != nil {
		http.Error(w, "Error parseando JSON", http.StatusBadRequest)
		return
	}

	mensaje.Evento = strings.ToLower(mensaje.Evento)
	global.LoggerKernel.Log("Kernel recibió evento de IO: "+mensaje.NombreIO+" ("+mensaje.Evento+")", log.DEBUG)

	switch mensaje.Evento {
	case "registro":
		//TODO registrarNuevoIO(mensaje)
	case "fin":
		//TODO conexionConIO.finalizarIO(mensaje)
	case "desconexion":
		//TODO conexionConIO.desconectarIO(mensaje)
	default:
		global.LoggerKernel.Log("Evento IO desconocido: "+mensaje.Evento, log.ERROR)
		http.Error(w, "Evento desconocido", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
