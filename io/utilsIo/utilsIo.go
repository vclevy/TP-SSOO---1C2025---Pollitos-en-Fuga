package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/io/global"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type PaqueteHandshakeIO = estructuras.PaqueteHandshakeIO

func HandshakeConKernel(paquete PaqueteHandshakeIO) error {
	global.LoggerIo.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), log.DEBUG)

	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerIo.Log("Error codificando paquete a JSON: "+err.Error(), log.ERROR)
		return err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/handshakeIO", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel, err.Error()), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerIo.Log("Respuesta HTTP del Kernel: "+resp.Status, log.DEBUG)

	return nil
}

func IniciarIo(solicitud estructuras.TareaDeIo) {
	// Log de inicio de E/S
	global.LoggerIo.Log(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %dms", solicitud.PID, solicitud.TiempoEstimado), log.INFO)

	// Simulación del proceso de E/S con sleep
	time.Sleep(time.Duration(solicitud.TiempoEstimado) * time.Millisecond)

	InformarFinalizacionDeIO(solicitud.PID)
}

func InformarFinalizacionDeIO(pid int){
	
	url := fmt.Sprintf("http://%s:%d/finalizacionIO?pid=%d", global.IoConfig.IPKernel, global.IoConfig.Port_Kernel,pid)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	pidString := strconv.Itoa(pid)
	global.LoggerIo.Log(fmt.Sprintf("## PID: <%s> - Fin de IO", pidString), log.INFO)
}

