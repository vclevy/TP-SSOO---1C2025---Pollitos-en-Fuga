package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"io"
)

func MemoriaLee(direccionFisica int, tamanio int) (string ,error) {
	datosEnvio := estructuras.PedidoREAD{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Tamanio:         tamanio,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return "" , fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/leerMemoria", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando pedido lectura a Memoria: "+err.Error(), log.ERROR)
		return "" , err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pedido lectura fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Pedido lectura enviado a Memoria con éxito", log.INFO)

	body, _ := io.ReadAll(resp.Body)

	var contenido string
	err = json.Unmarshal(body, &contenido)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return "" , err
	}

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, contenido), log.INFO) //!! Lectura/Escritura Memoria - logObligatorio


	return contenido , nil
}

func MemoriaEscribe(direccionFisica int, datos string) error {

	datosEnvio := estructuras.PedidoWRITE{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Datos:           []byte(datos),
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

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, datos), log.INFO) //!! Lectura/Escritura Memoria - logObligatorio

	return nil
}

func MemoriaEscribePaginaCompleta(direccionFisica int, datos []byte) error {
	datosEnvio := estructuras.PedidoWRITE{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Datos:      	 datos,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/actualizarPaginaCompleta", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

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

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, datos), log.INFO) //!! Lectura/Escritura Memoria (página completa) - logObligatorio


	return nil
}

func MemoriaLeePaginaCompleta(direccionFisica int) (string ,error) {
	datosEnvio := estructuras.PedidoREAD{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Tamanio:         0,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return "" , fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/leerPaginaCompleta", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando pedido lectura a Memoria: "+err.Error(), log.ERROR)
		return "" , err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pedido lectura fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Pedido lectura enviado a Memoria con éxito", log.INFO)

	body, _ := io.ReadAll(resp.Body)

	var contenido string
	err = json.Unmarshal(body, &contenido)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return "" , err
	}

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, contenido), log.INFO)  //!! Lectura/Escritura Memoria (página completa) - logObligatorio


	return contenido , nil
}
