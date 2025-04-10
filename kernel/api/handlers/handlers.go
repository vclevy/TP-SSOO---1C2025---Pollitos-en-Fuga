package handlers

import (
	"fmt"
	"net/http"
	//logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func SaludoKernel(w http.ResponseWriter, r *http.Request) { // w http.ResponseWriter, r *http.Request
	// loggerSaludo := logger.ConfigurarLogger("app.log", "dev")
	// loggerSaludo.Log("Hola", logger.DEBUG)
	fmt.Println("KERNEL RECIBIO UNA SOLICITUD DE /HOLA")
}
