package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	utilsMemoria"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	

)

type PaqueteMemoria = estructuras.PaqueteMemoria
type PaqueteSolicitudInstruccion = estructuras.SolicitudInstruccion
type PaqueteConfigMMU = estructuras.ConfiguracionMMU
type AccesoTP = estructuras.AccesoTP
type PedidoREAD = estructuras.PedidoREAD

//el KERNEL manda un proceso para inicializar con la estrcutura de PaqueteMemoria
func InicializarProceso(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var paquete PaqueteMemoria
    if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
        http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
        return
    }

	pid := paquete.PID
	tamanio := paquete.TamanioProceso
    archivoPseudocodigo := paquete.ArchivoPseudocodigo
	
	pidString := strconv.Itoa(pid)

	utilsMemoria.CrearTablaPaginas(pid, tamanio)
	utilsMemoria.ReservarMarcos(pid, tamanio)
	utilsMemoria.CargarProceso(pid, archivoPseudocodigo)
	global.LoggerMemoria.Log("## "+ pidString +": <"+ pidString +"> - Proceso Creado - Tamaño: <"+strconv.Itoa(tamanio)+">", log.DEBUG)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Paquete recibido correctamente para PID %d", paquete.PID)
}

//e KERNEL comprueba que haya espacio disponible en memoria antes de inicializar
func VerificarEspacioDisponible(w http.ResponseWriter, r *http.Request) {
	tamanioStr := r.URL.Query().Get("tamanio") // http/ip:puerto/verificarEspacioDisponoble?verificarEspacioDisponoble=432
	
	// Intentamos convertir el parámetro a entero
	tamanio, err := strconv.Atoi(tamanioStr)
	if err != nil {
		http.Error(w, "Tamaño de proceso inválido", http.StatusBadRequest)
		return
	}
	// Verificamos si hay suficiente espacio en la memoria
	espacioDisponible := utilsMemoria.HayLugar(tamanio)
	if espacioDisponible {
		// Si hay espacio, respondemos con un OK
		w.WriteHeader(http.StatusOK)
	} else {
		// Si no hay espacio, respondemos con un error
		http.Error(w, "No hay suficiente espacio", http.StatusConflict)
	}

	// Retornamos la respuesta en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"espacioDisponible": espacioDisponible})
}

//la CPU pide una instruccion
func DevolverInstruccion(w http.ResponseWriter, r *http.Request) {
	
	var paquete PaqueteSolicitudInstruccion

	// Decodificar JSON recibido
	err := json.NewDecoder(r.Body).Decode(&paquete)
	if err != nil {
		http.Error(w, "Error al parsear la solicitud", http.StatusBadRequest)
		return
	}

	// Obtener instrucción desde memoria
	instruccion, err := utilsMemoria.ObtenerInstruccion(paquete.Pid, paquete.Pc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Devolver instrucción como JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(instruccion); err != nil {
		http.Error(w, "Error al enviar la instrucción", http.StatusInternalServerError)
		return
	}
	pidString := strconv.Itoa(paquete.Pid)
	pcString :=  strconv.Itoa(paquete.Pc)
	
	global.LoggerMemoria.Log("## "+ pidString +": <"+ pidString +"> - Obtener instrucción: <"+ pcString +"> - Instrucción: <"+ instruccion +"> <...ARGS>", log.DEBUG)
}

//CPU lo pide
func ArmarPaqueteConfigMMU(w http.ResponseWriter, r *http.Request) {
	paquete := PaqueteConfigMMU {
			Tamanio_pagina :global.ConfigMemoria.Page_Size,
			Cant_entradas_tabla : global.ConfigMemoria.Entries_per_page,
			Cant_N_Niveles: global.ConfigMemoria.Number_of_levels,	
	} 

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(paquete); err != nil {
		http.Error(w, "Error al enviar la configuracion a CPU sobre MMU", http.StatusInternalServerError)
		return
	}
}

func AccederTablaPaginas(w http.ResponseWriter, r *http.Request) {
	//llega un PID y direccion logica
	//hace la traduccion
	//devuelve la direccion fisica 
	//++ metricas de Acceso a TP
	if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var paquete AccesoTP
    if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
        http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
        return
    }

	pid := paquete.PID
	direccionLogica := paquete.DireccionLogica

	utilsMemoria.TraducirLogicaAFisica(pid, direccionLogica)

}

//ACESSO A ESPACIO DE USUARIO
func LeerMemoria(w http.ResponseWriter, r *http.Request) {
    // input: pid, direccion_fisica, tamaño
    // output: contenido
	if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var paquete PedidoREAD
    if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
        http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
        return
    }
	pid := paquete.PID
	direccionFisica := paquete.DireccionFisica
	tamanio := paquete.Tamanio

	//faltaria validar q el tamanio sea valido
	datos := utilsMemoria.DevolverLecturaMemoria(pid, direccionFisica, tamanio)
	
	if err := json.NewEncoder(w).Encode(datos); err != nil {
		http.Error(w, "Error al enviar la instrucción", http.StatusInternalServerError)
		return
	}
}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
    // input: pid, direccion_fisica, datos
    // output: OK o error
}

func AccederEspacioUsuario(w http.ResponseWriter, r *http.Request){
	//ante un pedido de lectura, devolver el valor de esa posicion
	//ante pedido de escrita escribir lo pedido
	// PENSAR si conviene hacer una para write y otra para read
	//en ambos casos edita las metricas
}

//KERNEL notifica a memoria que finalizo
func FinalizarProceso(w http.ResponseWriter, r *http.Request){
	//libera su espacio en memoria y marcar como libres sus entradas en SWAP
	//genera log con las metricas

}
