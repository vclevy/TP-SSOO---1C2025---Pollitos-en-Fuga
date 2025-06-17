package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func MemoriaLee(direccionFisica int, tamanio int) error {
	datosEnvio := estructuras.PedidoREAD{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Tamanio:         tamanio,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/leerMemoria", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando pedido lectura a Memoria: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pedido lectura fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Pedido lectura enviado a Memoria con éxito", log.INFO)

	return nil
}

func MemoriaEscribe(direccionFisica int, datos string) error {
	datosEnvio := estructuras.PedidoWRITE{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Datos:           datos,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/escribirMemoria", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando pedido escritura a Memoria: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pedido escritura fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Pedido escritura enviados a Memoria con éxito", log.INFO)

	return nil
}

func MemoriaActualiza(direccionFisica int, datos string) error {
	datosEnvio := estructuras.PedidoWRITE{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Datos:           datos,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/escribirMemoria", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //!!cambiar

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando pedido escritura a Memoria: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pedido escritura fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Pedido escritura enviados a Memoria con éxito", log.INFO)

	return nil
}