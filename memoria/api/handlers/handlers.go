package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
	utilsMemoria"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	

)

type PaqueteMemoria = estructuras.PaqueteMemoria

func RecibirProceso(w http.ResponseWriter, r *http.Request) {
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
	
	utilsMemoria.CargarProceso(pid, archivoPseudocodigo)
	global.LoggerMemoria.Log("## "+ pidString +": <"+ pidString +"> - Proceso Creado - Tamaño: <"+strconv.Itoa(tamanio)+">", log.DEBUG)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Paquete recibido correctamente para PID %d", paquete.PID)
}




func VerificarEspacioDisponible(w http.ResponseWriter, r *http.Request) {
	tamanioStr := r.URL.Query().Get("tamanioProceso") // Aquí debes asegurarte que el parámetro esté correcto
	
	// Intentamos convertir el parámetro a entero
	tamanio, err := strconv.Atoi(tamanioStr)
	if err != nil {
		http.Error(w, "Tamaño de proceso inválido", http.StatusBadRequest)
		return
	}
//arreglar nombres de funciones
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

