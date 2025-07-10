package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var instruccionesConMMU = map[string]bool{
	"WRITE":       true,
	"READ":        true,
	"IO":          false,
	"BLOCKED":     false,
	"INIT_PROC":   false,
	"DUMP_MEMORY": false,
	"EXIT":        false,
	"NOOP":        false,
	"GOTO":        false,
}

type Instruccion struct {
	Opcode     string   `json:"opcode"`     // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción
}


var direccionFisica int

var nroPagina int
var Rafaga int

func HandshakeKernel() error {
	datosEnvio := estructuras.HandshakeConCPU{
		ID:     global.CpuID,
		Puerto: global.CpuConfig.Port_Cpu,
		IP:     global.CpuConfig.Ip_Cpu,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando handshake: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/handshakeCPU", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando handshake al Kernel: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("handshake fallido con status %d", resp.StatusCode)
	}

	return nil
}

func instruccionAEjecutar(solicitudInstruccion estructuras.PCB) string {

	jsonData, err := json.Marshal(solicitudInstruccion)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return ""
	}

	url := fmt.Sprintf("http://%s:%d/solicitudInstruccion", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                                          //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: "+err.Error(), log.ERROR)
		return ""
	}
	defer resp.Body.Close() //se cierra la conexión

/* 	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria de forma exitosa", log.INFO)
 */
	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var instruccionAEjecutar string
	err = json.Unmarshal(body, &instruccionAEjecutar)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return ""
	}
	return instruccionAEjecutar
}

func cortoProceso() error {
	datosEnvio := estructuras.RespuestaCPU{
		PID:        global.PCB_Actual.PID,
		PC:         global.PCB_Actual.PC,
		Motivo:     global.Motivo,
		RafagaReal: global.Rafaga,
	}

	jsonData, err := json.Marshal(datosEnvio)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/devolucion", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error devolviendo proceso a Kernel: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("devolución proceso fallido con status %d", resp.StatusCode)
	}
	global.LoggerCpu.Log("✅ Devolución proceso enviado a Kernel con éxito", log.INFO)
	return nil
}

func Desalojo() {
	// Solo vaciar el PCB si el proceso finalizó (EXIT) o fue interrumpido (READY)
	
	if global.CacheHabilitada {
		for i := 0; i < global.CpuConfig.CacheEntries; i++ {
			if global.CACHE[i].BitModificado == 1 {
				nroPaginaDesalojar := global.CACHE[i].NroPagina
				desalojar(i, nroPaginaDesalojar)
			}

			global.CACHE[i].BitModificado = -1
			global.CACHE[i].NroPagina = -1
			global.CACHE[i].Contenido = make([]byte, global.ConfigMMU.Tamanio_pagina)
			global.CACHE[i].BitUso = -1
		}
	}

	if global.TlbHabilitada {
		for i := 0; i < global.CpuConfig.CacheEntries; i++ {
			if global.TLB[i].NroPagina != -1 {
				global.TLB[i].NroPagina = -1
				global.TLB[i].Marco = -1
				global.TLB[i].UltimoUso = 0
			}
		}
	}
	if global.Motivo == "EXIT" || global.Motivo == "READY" {
		global.PCB_Actual = nil
	}
}

func desalojar(indicePisar int, nroPaginaPisar int) {
	marco := CalcularMarco(nroPaginaPisar)
	direccionFisica := MMU(0, marco)

	MemoriaEscribePaginaCompleta(direccionFisica, global.CACHE[indicePisar].Contenido)
	global.LoggerCpu.Log(fmt.Sprintf("\033[36mPID: %d - Memory Update - Página: %d - Frame: %d\033[0m", global.PCB_Actual.PID, global.CACHE[indicePisar].NroPagina, marco), log.INFO) //!! Página Actualizada de Caché a Memoria - LogObligatorio
}

func DevolucionPID() error {
	response := estructuras.RespuestaCPU{
		PID:        global.PCB_Actual.PID,
		PC:         global.PCB_Actual.PC,
		Motivo:     global.Motivo,
		RafagaReal: global.Rafaga,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error codificando pedido: %w", err)
	}
	url := fmt.Sprintf("http://%s:%d/devolucion", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		global.LoggerCpu.Log("Error enviando devolucióna Kernel: "+err.Error(), log.ERROR)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("devolución a kernel fallida con status %d", resp.StatusCode)
	}
	/* global.LoggerCpu.Log("✅ Devolución de PID enviado a Kernel con éxito", log.INFO) */

	return nil
}

