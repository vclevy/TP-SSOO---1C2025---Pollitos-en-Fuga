package handlers

import (
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func EscribirKernel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Solo se acepta POST", http.StatusMethodNotAllowed)
		global.LoggerKernel.Log("Método no permitido en /escribir", log.ERROR)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
		global.LoggerKernel.Log("Error leyendo cuerpo: "+err.Error(), log.ERROR)
		return
	}

	msg := string(body)
	global.LoggerKernel.Log("Mensaje recibido: "+msg, log.DEBUG) //memoria
	w.Write([]byte("Kernel recibió el mensaje")) //kernel
}