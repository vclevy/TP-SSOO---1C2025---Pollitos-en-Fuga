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
	"strconv"
	"strings"
	"time"
)

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
var desplazamiento int
var nroPagina int
var Marco int
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

func Fetch() string {
	pidActual := global.PCB_Actual.PID
	pcActual := global.PCB_Actual.PC

	global.LoggerCpu.Log(fmt.Sprintf(" ## PID: %d - FETCH - Program Counter: %d", pidActual, pcActual), log.INFO)

	solicitudInstruccion := estructuras.PCB{
		PID: pidActual,
		PC:  pcActual,
	}

	var instruccionAEjecutar = instruccionAEjecutar(solicitudInstruccion)

	global.LoggerCpu.Log(fmt.Sprintf("Memoria respondió con la instrucción: %s", instruccionAEjecutar), log.INFO)

	return instruccionAEjecutar
}

func Decode(instruccionAEjecutar string) (Instruccion, bool) {
	instruccionPartida := strings.Fields(instruccionAEjecutar) //?  "MOV AX BX" --> []string{"MOV", "AX", "BX"}

	instruccion := Instruccion{
		Opcode:     instruccionPartida[0],
		Parametros: instruccionPartida[1:],
	}

	_, requiereMMU := instruccionesConMMU[instruccion.Opcode]

	return instruccion, requiereMMU
}

func Execute(instruccion Instruccion, requiereMMU bool) error {

	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", global.PCB_Actual.PID, instruccion.Opcode, instruccion.Parametros), log.INFO)

	//todo INSTRUCCIONES SYSCALLS
	if instruccion.Opcode == "IO" {
		global.Motivo = "BLOCKED"
		cortoProceso()
		Syscall_IO(instruccion)
	}
	if instruccion.Opcode == "INIT_PROC" {
		Syscall_Init_Proc(instruccion)
	}
	if instruccion.Opcode == "DUMP_MEMORY" {
		Syscall_Dump_Memory()
	}
	if instruccion.Opcode == "EXIT" {
		global.Motivo = "EXIT"
		cortoProceso()
		Syscall_Exit()
	}

	//todo OTRAS INSTRUCCIONES
	if instruccion.Opcode == "NOOP" {
	}

	if instruccion.Opcode == "GOTO" {
		pcNuevo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			return fmt.Errorf("error al convertir tiempo estimado")
		}
		global.PCB_Actual.PC = pcNuevo
	}

	//todo INSTRUCCIONES MMU
	if requiereMMU {
		tlbHabilitada := global.CpuConfig.TlbEntries > 0
		cacheHabilitada := global.CpuConfig.CacheEntries > 0

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			return fmt.Errorf("error al convertir dirección logica")
		} else {
			ConfigMMU()
			desplazamiento = direccionLogica % configMMU.Tamanio_pagina
			nroPagina = direccionLogica / configMMU.Tamanio_pagina
		}

		if instruccion.Opcode == "READ" { // READ 0 20 - READ (Dirección, Tamaño)
			/* 			Read(instruccion, cacheHabilitada, tlbHabilitada, direccionLogica)
			 */
		}

		if instruccion.Opcode == "WRITE" { // WRITE 0 EJEMPLO_DE_ENUNCIADO - WRITE (Dirección, Datos)
			WRITE(instruccion, cacheHabilitada, tlbHabilitada, direccionLogica)
		}
	}

	return nil
}

/* func CacheHIT() bool {
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == nroPagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! CACHE HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! CACHE MISS
	return false
}
*/
func TlbHIT() bool {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == nroPagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! TLB HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!!TLB MISS
	return false
} 

func CheckInterrupt() {
	if global.Interrupcion {
		global.Motivo = "READY"
		cortoProceso()
		global.PCB_Actual = global.PCB_Interrupcion
		global.Interrupcion = false
	}
}


func WRITE(instruccion Instruccion, cacheHabilitada bool, tlbHabilitada bool, direccionLogica int) {
	dato := instruccion.Parametros[1]
	if cacheHabilitada {
		actualizarCACHE(nroPagina, dato)
	} else {
		if tlbHabilitada {
			var marco int
			if TlbHIT() {
				marco = global.TLB[indice].Marco
			} else {
				marco = CalcularMarco()
			}
			direccionFisica = MMU(direccionLogica, instruccion.Opcode, nroPagina, marco)
			MemoriaEscribe(direccionFisica, dato)
			actualizarTLB(nroPagina, marco)
		} else {
			marco := CalcularMarco()
			direccionFisica = MMU(direccionLogica, instruccion.Opcode, nroPagina, marco)
			MemoriaEscribe(direccionFisica, dato)
		}
	}
}

func actualizarCACHE(pagina int, nuevoContenido string) {
	indice := indicePagina(pagina)
	if indice == -1 { 
		global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! CACHE MISS
		paginaPisar := AlgoritmoCACHE()
		indicePisar := indicePagina(paginaPisar)
		if global.CACHE[indicePisar].BitModificado == 1 {
			MemoriaEscribe(direccionFisica, global.CACHE[indicePisar].Contenido) //!! ver dirección fisica
		}
		global.CACHE[indicePisar].NroPagina = pagina
		global.CACHE[indicePisar].Contenido = nuevoContenido
		global.CACHE[indicePisar].BitModificado = 0
	} else {
		global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! CACHE HIT	
		global.CACHE[indice].Contenido = nuevoContenido
		global.CACHE[indice].BitModificado = 1
	}
}

func actualizarTLB(pagina int, marco int) {
	indice := indicePagina(pagina)
	if indice == -1 { // no está la página
		paginaPisar := AlgoritmoTLB()
		indicePisar := indicePagina(paginaPisar)
		global.TLB[indicePisar].Marco = marco
		global.TLB[indicePisar].NroPagina = pagina
	} else {
		global.TLB[indice].Marco = marco
	}
}

func indicePagina(pagina int) int {
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			return i
		}
	}
	return -1
}

/* func Read(instruccion Instruccion, cacheHabilitada bool, tlbHabilitada bool, direccionLogica int) error {
	tamanioStr := instruccion.Parametros[1]
	tamanio, err := strconv.Atoi(tamanioStr)
	if err != nil {
		return fmt.Errorf("error al convertir tamanio")
	}

	if cacheHabilitada && CacheHIT(){
		global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: %s - Dirección Física: %d - Valor: %s.", global.PCB_Actual.PID, instruccion.Opcode, direccionFisica, global.CACHE[indice].Contenido), log.INFO)
	}else if  tlbHabilitada && TlbHIT() {
			marco := global.TLB[indice].Marco
			direccionFisica = MMU(direccionLogica, instruccion.Opcode, nroPagina, marco)
			MemoriaLee(direccionFisica, tamanio)
	} else {
		marco := CalcularMarco()
		direccionFisica = MMU(direccionLogica, instruccion.Opcode, nroPagina, marco)
		MemoriaLee(direccionFisica, tamanio)
	}
	return nil
} */


/*
LOGS:
//Fetch Instrucción: “## PID: <PID> - FETCH - Program Counter: <PROGRAM_COUNTER>”.
//Interrupción Recibida: “## Llega interrupción al puerto Interrupt”.
//Instrucción Ejecutada: “## PID: <PID> - Ejecutando: <INSTRUCCION> - <PARAMETROS>”.
Lectura/Escritura Memoria: “PID: <PID> - Acción: <LEER / ESCRIBIR> - Dirección Física: <DIRECCION_FISICA> - Valor: <VALOR LEIDO / ESCRITO>”.
//Obtener Marco: “PID: <PID> - OBTENER MARCO - Página: <NUMERO_PAGINA> - Marco: <NUMERO_MARCO>”.
//TLB Hit: “PID: <PID> - TLB HIT - Pagina: <NUMERO_PAGINA>”
//TLB Miss: “PID: <PID> - TLB MISS - Pagina: <NUMERO_PAGINA>”
//Página encontrada en Caché: “PID: <PID> - Cache Hit - Pagina: <NUMERO_PAGINA>”
//Página faltante en Caché: “PID: <PID> - Cache Miss - Pagina: <NUMERO_PAGINA>”
Página ingresada en Caché: “PID: <PID> - Cache Add - Pagina: <NUMERO_PAGINA>” //?? la pagina que reemplaza a otra, según el algoritmo
Página Actualizada de Caché a Memoria: “PID: <PID> - Memory Update - Página: <NUMERO_PAGINA> - Frame: <FRAME_EN_MEMORIA_PRINCIPAL>” //?? la pagina que se desaloja, según el algoritmo
*/
