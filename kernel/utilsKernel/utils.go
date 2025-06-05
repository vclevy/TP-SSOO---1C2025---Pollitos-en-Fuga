package utilskernel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

func FiltrarColaIO(cola []*ProcesoIO, target *ProcesoIO) []*ProcesoIO {
	resultado := make([]*ProcesoIO, 0, len(cola))
	for _, pio := range cola {
		if pio != target {
			resultado = append(resultado, pio)
		}
	}
	return resultado
}

func FiltrarIODevice(lista []*IODevice, excluir *IODevice) []*IODevice {
	var resultado []*IODevice
	for _, io := range lista {
		if io != excluir {
			resultado = append(resultado, io)
		}
	}
	return resultado
}

func SolicitarDumpAMemoria(pid int) error {
	url := fmt.Sprintf("http://%s:%d/dump", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory)

	body := estructuras.SolicitudDump{PID: pid}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("fallo conexión con Memoria: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("memoria devolvió error: código %d", resp.StatusCode)
	}

	return nil
}

func BuscarCPUPorPID(pid int) *global.CPU {
    global.MutexCPUs.Lock()
    defer global.MutexCPUs.Unlock()
    for _, cpu := range global.CPUsConectadas {
        if cpu.ProcesoEjecutando != nil && cpu.ProcesoEjecutando.PID == pid {
            return cpu
        }
    }
    return nil
}


func EnviarADispatch(cpu *global.CPU, pid int, pc int) (*estructuras.RespuestaCPU, error) {
	url := fmt.Sprintf("http://%s:%d/dispatch", cpu.IP, cpu.Puerto)

	payload := map[string]interface{}{
		"pid": pid,
		"pc":  pc,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error serializando payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error enviando request HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta no OK del dispatch: %d", resp.StatusCode)
	}

	var respuesta estructuras.RespuestaCPU
	if err := json.NewDecoder(resp.Body).Decode(&respuesta); err != nil {
		return nil, fmt.Errorf("error parseando respuesta JSON: %w", err)
	}

	return &respuesta, nil
}


func EnviarInterrupcionCPU(cpu *global.CPU, pid int) (error) {
	url := fmt.Sprintf("http://%s:%d/interrupt", cpu.IP, cpu.Puerto)

	payload := map[string]interface{}{
		"pid": pid,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error enviando request HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("respuesta no OK del interrupt: %d", resp.StatusCode)
	}

	// Leer respuesta
	var response struct {
		PID int `json:"pid"`
		PC  int `json:"pc"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return  fmt.Errorf("error decodificando respuesta: %w", err)
	}

	return nil
}



func HayCPUDisponible() bool {
	for _, cpu := range global.CPUsConectadas {
		if cpu.ProcesoEjecutando == nil {
			return true
		}
	}
	return false
}

func SolicitarMemoria(tamanio int) bool {
	cliente := &http.Client{}
	endpoint := "verificarEspacioDisponible"
	url := fmt.Sprintf("http://%s:%d/%s?tamanio=%d", global.ConfigKernel.IPMemory, global.ConfigKernel.Port_Memory, endpoint, tamanio)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error creando request para solicitar memoria: %v", err), log.ERROR)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(req)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error enviando request para solicitar memoria: %v", err), log.ERROR)
		return false
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		global.LoggerKernel.Log(fmt.Sprintf("Memoria respondió con status %d para solicitud de %d bytes", respuesta.StatusCode, tamanio), log.ERROR)
		return false
	}

	return true
}

func MoverASwap(pid int) error {
	url := fmt.Sprintf("http://%s:%d/moverASwap?pid=%d",
		global.ConfigKernel.IPMemory,
		global.ConfigKernel.Port_Memory,
		pid)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor de memoria")
	}
	return nil
}

func MoverAMemoria(pid int) error {
	url := fmt.Sprintf("http://%s:%d/moverAMemoria?pid=%d",
		global.ConfigKernel.IPMemory,
		global.ConfigKernel.Port_Memory,
		pid)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la respuesta del servidor de memoria")
	}
	return nil
}

func InformarFinAMemoria(pid int) error {
	url := "http://" + global.ConfigKernel.IPMemory + ":" + strconv.Itoa(global.ConfigKernel.Port_Memory) + "/finalizarProceso"
	data := map[string]int{"pid": pid}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("memoria devolvió error")
	}
	return nil
}

func BuscarProcesoPorPID(cola []*global.Proceso, pid int) (*global.Proceso) {
	for i := range cola {
		if cola[i].PCB.PID == pid {
			return cola[i]
		}
	}
	return nil
}
