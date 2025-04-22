package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	utilsKernel "github.com/sisoputnfrba/tp-golang/kernel/utilsKernel"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	utils "github.com/sisoputnfrba/tp-golang/utils/paquetes"
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

type PCB = utilsKernel.PCB

func NuevoPCB(pid int) *PCB {
	return &PCB{
		PID: pid,
		PC:  0,
		ME:  make(map[string]int),
		MT:  make(map[string]int),
	}
}

type Proceso struct {
	PCB
	MemoriaRequerida int
}


// Vamos a necesitar aca una api con w*responseWritter y eso para el handler que contiene la func crear proceso


func INIT_PROC(w http.ResponseWriter, r *http.Request){
	//? archivo := r.URL.Query().Get("archivo") //? no se que hacer con el archivo este
	tamanioStr := r.URL.Query().Get("tamanio")

	str_pid := utils.LeerStringDeConsola()
	pid, _ := strconv.Atoi(str_pid)
	pcb := NuevoPCB(pid)

	tamanio, err := strconv.Atoi(tamanioStr)
	if err != nil {
		http.Error(w, "Error al convertir el tamaño a entero", http.StatusBadRequest)
		global.LoggerKernel.Log("Error al convertir el tamaño a entero: "+err.Error(), log.DEBUG)
		return
	}

	procesoCreado := Proceso{PCB: *pcb, MemoriaRequerida: tamanio}
	global.LoggerKernel.Log(fmt.Sprintf("Proceso creado: %+v", procesoCreado), log.DEBUG)

	global.ColaNew = append(global.ColaNew, global.Proceso(procesoCreado)) // no estoy segura si esta bien la sintaxis
}