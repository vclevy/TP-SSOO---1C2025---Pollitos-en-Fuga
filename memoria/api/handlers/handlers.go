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

)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int  `json:"codigo"`
}

func RecibirPaquete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		global.LoggerMemoria.Log("Se intentó acceder con un método no permitido", log.DEBUG)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerMemoria.Log("Error leyendo el cuerpo del request: "+err.Error(), log.DEBUG)
		return
	}
	defer r.Body.Close()

	var paquete Paquete
	err = json.Unmarshal(body, &paquete)
	if err != nil {
		http.Error(w, "Error parseando el paquete", http.StatusBadRequest)
		global.LoggerMemoria.Log("Error al parsear el paquete JSON: "+err.Error(), log.DEBUG)
		return
	}
	global.LoggerMemoria.Log("Memoria recibió paquete: Mensajes: "+strings.Join(paquete.Mensajes, ", ")+" Codigo: "+strconv.Itoa(paquete.Codigo), log.DEBUG)

	w.Write([]byte("Memoria recibió el paquete correctamente"))
}

func TamanioProceso(w http.ResponseWriter, r *http.Request) {
    partes := strings.Split(r.URL.Path, "/")
    if len(partes) != 2 {
        http.Error(w, "Faltan parámetros", http.StatusBadRequest)
        return
    }

    tamanio, err := strconv.Atoi(partes[1])
    if err != nil {
        http.Error(w, "Tamaño inválido", http.StatusBadRequest)
        return
    }

    fmt.Printf("Me pidieron reservar %d bytes\n", tamanio)
    w.WriteHeader(http.StatusOK) // o algún código según si tenés espacio o no

	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utilsMemoria.VerificarEspacioDisponible(tamanio))
}