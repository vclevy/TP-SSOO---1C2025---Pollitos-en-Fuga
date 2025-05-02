package utilskernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type IoDevice = global.IODevice

func ObtenerDispositivoIO(nombreBuscado string) []*global.IODevice {
    var dispositivos []*global.IODevice
    for _, io := range global.IOConectados {
        if io.Nombre == nombreBuscado {
            dispositivos = append(dispositivos, io)
        }
    }
    return dispositivos
}

func EnviarAIO(dispositivo *IoDevice, pid int, tiempoUso int){
    puerto := dispositivo.Puerto
    ip := dispositivo.IP

	paqueteAEnviar := estructuras.TareaDeIo{
		PID:           pid,
		TiempoEstimado: tiempoUso,
	}

	jsonData, _ := json.Marshal(paqueteAEnviar)
    url := fmt.Sprintf("http://%s:%d/procesoRecibido", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerKernel.Log("Error enviando handshake al Kernel: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close()

}
