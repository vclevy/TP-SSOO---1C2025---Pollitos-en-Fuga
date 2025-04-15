package handlers

import (
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Solo se acepta POST", http.StatusMethodNotAllowed)
		global.Logger.Log("Método no permitido en /escribir", log.ERROR)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.Logger.Log("Error leyendo cuerpo: "+err.Error(), log.ERROR)
		return
	}

	msg := string(body)
	global.Logger.Log("Mensaje recibido: "+msg, log.DEBUG) //memoria
	w.Write([]byte("Memoria recibió el mensaje")) //kernel
}