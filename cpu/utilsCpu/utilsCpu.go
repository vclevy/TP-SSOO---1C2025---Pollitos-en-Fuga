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
	"time"
)

var lruCounter int
var lruValues []int // lruValues tiene el timestamp de cada marco de TLB

var instruccionesConMMU = map[string]bool{
	"WRITE": true,
	"READ":  true,
}

type Instruccion struct {
	Opcode     string   `json:"opcode"`     // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción, de tipo variable
}

var configMMU estructuras.ConfiguracionMMU
var direccionFisica int

var nroPagina int
var indice int
var Rafaga int
var tiempoInicio time.Time

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
	global.Rafaga = time.Since(tiempoInicio).Seconds()

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

func CicloDeInstruccion() {
	var instruccionAEjecutar = Fetch()

	instruccion, requiereMMU := Decode(instruccionAEjecutar)

	tiempoInicio = time.Now()
	Execute(instruccion, requiereMMU)

	CheckInterrupt()
}

func TlbHIT(pagina int) bool {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo TLB HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo TLB MISS
	return false
}

func CacheHIT(pagina int) bool {
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo CACHE HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo CACHE MISS
	return false
}

func actualizarCACHE(pagina int, nuevoContenido string) {
	indice := indicePaginaEnCache(pagina)
	if indice == -2 { // no está la página en cache
		indicePisar := AlgoritmoCACHE()
		if global.CACHE[indicePisar].BitModificado == 1 {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO)
			desalojar(indicePisar)
		}
		global.CACHE[indicePisar].NroPagina = pagina
		global.CACHE[indicePisar].Contenido = nuevoContenido
		global.CACHE[indicePisar].BitModificado = 0
	} else {
		global.CACHE[indice].Contenido = nuevoContenido
		global.CACHE[indice].BitModificado = 1
	}
}

func actualizarTLB(pagina int, marco int) {
	indice := indicePaginaEnTLB(pagina)
	if indice == -2 { // no está la página
		indicePisar := AlgoritmoTLB()
		lruCounter++
		global.TLB[indicePisar].UltimoUso = lruCounter
		global.TLB[indicePisar].Marco = marco
		global.TLB[indicePisar].NroPagina = pagina
	} else {
		global.TLB[indice].Marco = marco
	}
}

func indicePaginaEnCache(pagina int) int {
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			return i
		}
	}
	return -2
}

func indicePaginaEnTLB(pagina int) int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == pagina {
			global.TLB[i].UltimoUso = lruCounter

			return i
		}
	}
	return -2
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
