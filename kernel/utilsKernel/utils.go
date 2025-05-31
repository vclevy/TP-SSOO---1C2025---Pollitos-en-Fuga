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

func MandarProcesoACPU(pcb global.PCB, cpu *global.CPU){
	// Serializar el PCB a JSON
	body, err := json.Marshal(pcb)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error serializando PCB: %v", err), log.ERROR)
		return
	}

	// URL del endpoint de la CPU
	url := fmt.Sprintf("http://%s:%d/ejecutarProceso", cpu.IP, cpu.Puerto)

	// Crear el request HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error creando request para CPU %s: %v", cpu.ID, err), log.ERROR)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Enviar el request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error enviando proceso a CPU %s: %v", cpu.ID, err), log.ERROR)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		global.LoggerKernel.Log(fmt.Sprintf("CPU %s respondió con error al ejecutar proceso (status: %d)", cpu.ID, resp.StatusCode), log.ERROR)
		return
	}

	global.LoggerKernel.Log(fmt.Sprintf("Proceso PID <%d> enviado a CPU %s correctamente", pcb.PID, cpu.ID), log.INFO)

	// Actualizar CPU como ocupada
	cpu.ProcesoEjecutando = &pcb
}

func InterrumpirCPU(cpu *global.CPU) {
	url := fmt.Sprintf("http://%s:%d/interrumpir", cpu.IP, cpu.Puerto)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error creando request de interrupción a CPU %s: %v", cpu.ID, err), log.ERROR)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		global.LoggerKernel.Log(fmt.Sprintf("Error enviando interrupción a CPU %s: %v", cpu.ID, err), log.ERROR)
		return
	}
	defer resp.Body.Close()

	global.LoggerKernel.Log(fmt.Sprintf("Interrupción enviada a CPU %s", cpu.ID), log.INFO)
}