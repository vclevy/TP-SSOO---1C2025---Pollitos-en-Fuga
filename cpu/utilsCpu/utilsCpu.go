package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

var lruCounter int

var instruccionesConMMU = map[string]bool{
	"WRITE": true,
	"READ":  true,
}

type Instruccion struct {
	Opcode     string   `json:"opcode"`     // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción
}

var configMMU estructuras.ConfiguracionMMU
var direccionFisica int

var nroPagina int
var indice int
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
	global.LoggerCpu.Log("✅ Handshake enviado al Kernel con éxito", log.INFO)

	return nil
}

func instruccionAEjecutar(estructuras.PCB) string {
	var solicitudInstruccion estructuras.PCB

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

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria de forma exitosa", log.INFO)

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

func desalojar(indicePisar int) {
	var marco int
	if global.CpuConfig.TlbEntries > 0 && TlbHIT(global.CACHE[indicePisar].NroPagina) { // pagina en tlb
		marco = global.TLB[indicePisar].Marco
	} else {
		marco = CalcularMarco()
	}
	direccionFisica = marco * configMMU.Tamanio_pagina
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Memory Update - Página: %d - Frame: %d", global.PCB_Actual.PID, global.CACHE[indicePisar].NroPagina, marco), log.INFO) //?? la pagina que se desaloja, según el algoritmo

	MemoriaEscribe(direccionFisica, global.CACHE[indicePisar].Contenido)
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
	global.LoggerCpu.Log("✅ Devolución de PID enviado a Kernel con éxito", log.INFO)

	return nil
}

func InicializarEstructuras() {
	global.TLB = make([]estructuras.DatoTLB, global.CpuConfig.TlbEntries)
	for i := range global.TLB {
		global.TLB[i] = estructuras.DatoTLB{
			NroPagina: -1,
			Marco:     -1,
			UltimoUso: -1,
		}
	}

	global.CACHE = make([]estructuras.DatoCACHE, global.CpuConfig.CacheEntries)
	for i := range global.CACHE {
		global.CACHE[i] = estructuras.DatoCACHE{
			BitModificado: -1,
			NroPagina:     -1,
			Contenido:     "",
			BitUso:        -1,
		}
	}
}

/*
LOGS:
//Fetch Instrucción: “## PID: <PID> - FETCH - Program Counter: <PROGRAM_COUNTER>”.
//Interrupción Recibida: “## Llega interrupción al puerto Interrupt”.
//Instrucción Ejecutada: “## PID: <PID> - Ejecutando: <INSTRUCCION> - <PARAMETROS>”.
//Lectura/Escritura Memoria: “PID: <PID> - Acción: <LEER / ESCRIBIR> - Dirección Física: <DIRECCION_FISICA> - Valor: <VALOR LEIDO / ESCRITO>”.
//Obtener Marco: “PID: <PID> - OBTENER MARCO - Página: <NUMERO_PAGINA> - Marco: <NUMERO_MARCO>”.
//TLB Hit: “PID: <PID> - TLB HIT - Pagina: <NUMERO_PAGINA>”
//TLB Miss: “PID: <PID> - TLB MISS - Pagina: <NUMERO_PAGINA>”
//Página encontrada en Caché: “PID: <PID> - Cache Hit - Pagina: <NUMERO_PAGINA>”
//Página faltante en Caché: “PID: <PID> - Cache Miss - Pagina: <NUMERO_PAGINA>”
//Página ingresada en Caché: “PID: <PID> - Cache Add - Pagina: <NUMERO_PAGINA>” //?? la pagina que reemplaza a otra, según el algoritmo
//Página Actualizada de Caché a Memoria: “PID: <PID> - Memory Update - Página: <NUMERO_PAGINA> - Frame: <FRAME_EN_MEMORIA_PRINCIPAL>” //?? la pagina que se desaloja, según el algoritmo
*/
