package utilskernel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

type IODevice = global.IODevice
type Proceso = global.Proceso
type ProcesoIO = global.ProcesoIO

func ObtenerDispositivoIO(nombreBuscado string) []*global.IODevice {
	var dispositivos []*global.IODevice
	for _, io := range global.IOConectados {
		if io.Nombre == nombreBuscado {
			dispositivos = append(dispositivos, io)
		}
	}
	return dispositivos
}

func EnviarAIO(dispositivo *IODevice, pid int, tiempoUso int) {
	puerto := dispositivo.Puerto
	ip := dispositivo.IP

	paqueteAEnviar := estructuras.TareaDeIo{
		PID:            pid,
		TiempoEstimado: tiempoUso,
	}

	jsonData, _ := json.Marshal(paqueteAEnviar)
	url := fmt.Sprintf("http://%s:%d/procesoRecibido", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerKernel.Log("Error enviando el paquete a IO: "+err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close()

}

//@valenchu agregue estas funciones de abajo

func BuscarDispositivo(host string, port int) (*global.IODevice, error) {
	global.IOListMutex.RLock()
	defer global.IOListMutex.RUnlock()

	for _, dispositivo := range global.IOConectados {
		if dispositivo.IP == host && dispositivo.Puerto == port {
			return dispositivo, nil
		}
	}
	return nil, errors.New("dispositivo no encontrado")
}

func FiltrarCola(cola []*Proceso, target *Proceso) []*Proceso {
	resultado := make([]*Proceso, 0)
	for _, p := range cola {
		if p != target {
			resultado = append(resultado, p)
		}
	}
	return resultado
}

func FiltrarColaIO(cola []*ProcesoIO, target *ProcesoIO) []*ProcesoIO {
	resultado := make([]*ProcesoIO, 0, len(cola))
	for _, pio := range cola {
		if pio != target {
			resultado = append(resultado, pio)
		}
	}
	return resultado
}
