package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"path/filepath"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	utilsMemoria"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	myLogger "github.com/sisoputnfrba/tp-golang/utils/logger"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"log"
	"fmt"
)

type PaqueteMemoria = estructuras.PaqueteMemoria
type PaqueteSolicitudInstruccion = estructuras.PCB
type PaqueteConfigMMU = estructuras.ConfiguracionMMU
type AccesoTP = estructuras.AccesoTP
type PedidoREAD = estructuras.PedidoREAD
type PedidoWRITE = estructuras.PedidoWRITE

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
	ruta := filepath.Join(global.ConfigMemoria.Scripts_Path, archivoPseudocodigo)

	log.Printf("PID %d - archivo: '%s' - ruta: '%s'\n", pid, archivoPseudocodigo, ruta)
	espacioDisponible := utilsMemoria.HayLugar(tamanio)
	if !espacioDisponible {
	http.Error(w, "No hay suficiente espacio", http.StatusConflict)
	}

	w.WriteHeader(http.StatusOK)
	
	utilsMemoria.CrearTablaPaginas(pid, tamanio)
	utilsMemoria.CargarProceso(pid, ruta)
	utilsMemoria.InicializarMetricas(pid)
	global.LoggerMemoria.Log("## PID: "+ strconv.Itoa(pid) +"> - Proceso Creado - Tamaño: <"+strconv.Itoa(tamanio)+">", myLogger.INFO)
}

//KERNEL comprueba que haya espacio disponible en memoria antes de inicializar
func VerificarEspacioDisponible(w http.ResponseWriter, r *http.Request) {
	tamanioStr := r.URL.Query().Get("tamanio")
	tamanio,err := strconv.Atoi(tamanioStr)
	if err != nil {
		http.Error(w, "Tamano invalido", http.StatusConflict)
		}
		
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

func Suspender(w http.ResponseWriter, r *http.Request){
	pidStr := r.URL.Query().Get("suspension") 
	pid,err := strconv.Atoi(pidStr)

	if err != nil {
		http.Error(w, "PID invalido", http.StatusBadRequest)
		return
	}

	utilsMemoria.Suspender(pid)

}

func DesSuspender(w http.ResponseWriter, r *http.Request){
	
}

func FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	type FinalizarProcesoRequest struct {
		PID int `json:"pid"`
	}

	var paquete FinalizarProcesoRequest

	if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
		http.Error(w, "Error al parsear el cuerpo", http.StatusBadRequest)
		return
	}

	pid := paquete.PID

	stringMetricas := utilsMemoria.FinalizarProceso(pid)
	w.WriteHeader(http.StatusOK)
	global.LoggerMemoria.Log(stringMetricas, myLogger.INFO)
}

//la CPU pide una instruccion del diccionario de procesos
func DevolverInstruccion(w http.ResponseWriter, r *http.Request) {
	
	var paquete PaqueteSolicitudInstruccion

	// Decodificar JSON recibido
	err := json.NewDecoder(r.Body).Decode(&paquete)
	if err != nil {
		http.Error(w, "Error al parsear la solicitud", http.StatusBadRequest)
		return
	}
	pid := paquete.PID
	pc := paquete.PC

	// Obtener instrucción desde memoria
	instruccion, err := utilsMemoria.ObtenerInstruccion(pid, pc)
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
	
	global.LoggerMemoria.Log("## PID: "+ strconv.Itoa(pid) +">  - Obtener instrucción: <"+ strconv.Itoa(pc) +"> - Instrucción: <"+ instruccion +"> <...ARGS>",  myLogger.INFO)
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

//CPU pasa la direccion logica para que le devolvamos el marco
func AccederTablaPaginas(w http.ResponseWriter, r *http.Request) {
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
	entradas := paquete.Entradas

	marco := utilsMemoria.EncontrarMarco(pid, entradas)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(marco); err != nil {
		http.Error(w, "Error al enviar el marco", http.StatusInternalServerError)
		return
	}
}

//CPU queire leer o escribir en Espacio de usuario
func LeerMemoria(w http.ResponseWriter, r *http.Request) {
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

	if direccionFisica+tamanio > utilsMemoria.TamMemoria{
		http.Error(w, "Dirección física invalida", http.StatusBadRequest)
		return
	}
	datos := utilsMemoria.LeerMemoria(pid, direccionFisica, tamanio)
	
	if err := json.NewEncoder(w).Encode(datos); err != nil {
		http.Error(w, "Error al enviar la instrucción", http.StatusInternalServerError)
		return
	}
	global.LoggerMemoria.Log("## PID: <"+ strconv.Itoa(pid) +">- <Lectura> - Dir. Física: <"+ 
	strconv.Itoa(direccionFisica) +"> - Tamaño: <"+ strconv.Itoa(tamanio)+ "> ",  myLogger.INFO)

}

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var paquete PedidoWRITE
    if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
        http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
        return
    }

	pid := paquete.PID
	direccionFisica := paquete.DireccionFisica
	datos := paquete.Datos
	
	utilsMemoria.EscribirDatos(pid, direccionFisica, datos)

	w.WriteHeader(http.StatusOK)

	global.LoggerMemoria.Log(fmt.Sprintf("## PID: <%d> - <Escritura> - Dir. Física: <%d> - Datos: <%s> ", pid, direccionFisica,datos), myLogger.INFO) //!! Fetch Instrucción - logObligatorio

}


func LeerPaginaCompleta(w http.ResponseWriter, r *http.Request){
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
	
	utilsMemoria.LeerPaginaCompleta(pid, direccionFisica)

}

func EscribirPaginaCompleta(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    var paquete PedidoWRITE
    if err := json.NewDecoder(r.Body).Decode(&paquete); err != nil {
        http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
        return
    }

	pid := paquete.PID
	direccionFisica := paquete.DireccionFisica
	datos := paquete.Datos
	
	utilsMemoria.ActualizarPaginaCompleta(pid, direccionFisica, datos)
}


