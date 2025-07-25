package utilsIo

import (
	"bytes"
	"encoding/json"
	"strings"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
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
	Opcode     string   `json:"opcode"`
	Parametros []string `json:"parametros"`
}

var direccionFisica int
var nroPagina int

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
		RafagaReal: float64(time.Since(global.TiempoInicio).Milliseconds()),
		IO:         global.IO_Request,
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

	return nil
}

func Desalojo() {
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
		for i := 0; i < global.CpuConfig.TlbEntries; i++ {
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
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Memory Update - Página: %d - Frame: %d", global.PCB_Actual.PID, global.CACHE[indicePisar].NroPagina, marco), log.INFO) //!! Página Actualizada de Caché a Memoria - LogObligatorio
}

func DevolucionPID() error {
	response := estructuras.RespuestaCPU{
		PID:        global.PCB_Actual.PID,
		PC:         global.PCB_Actual.PC,
		Motivo:     global.Motivo,
		RafagaReal: float64(time.Since(global.TiempoInicio).Milliseconds()),
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
	
	return nil
}

/* 	func decodificarSiEsBase64(contenido []byte) string {
    s := strings.TrimRight(string(contenido), "\x00")

    decoded, err := base64.StdEncoding.DecodeString(s)
    if err == nil {
        return string(decoded)
    }

    return s
}	
 */
func MostrarContenido(dato []byte) string {
	s := strings.TrimRight(string(dato), "\x00")

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil && esTextoLegible(decoded) {
		return string(decoded)
	}

	if esTextoLegible([]byte(s)) {
		return s
	}

	return fmt.Sprintf("[binario] %x", dato)
}

func esTextoLegible(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			if b != '\n' && b != '\r' && b != '\t' {
				return false
			}
		}
	}
	return true
}
