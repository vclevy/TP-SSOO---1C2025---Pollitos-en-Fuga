package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var instruccionesConMMU = map[string]bool{
	"WRITE": true,
	"READ":  true,
}

type Instruccion struct {
	Opcode     string   `json:"opcode"`     // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción, de tipo variable
}

var direccionFisica int
var desplazamiento int
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

func CicloDeInstruccion() {
	var instruccionAEjecutar = Fetch()
	if instruccionAEjecutar == "FIN" {
		global.Motivo = "EXIT"
		return
	}
	instruccion, requiereMMU := Decode(instruccionAEjecutar)

	Execute(instruccion, requiereMMU)

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

func Execute(instruccion Instruccion, requiereMMU bool) {
	
	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", global.PCB_Actual.PID, instruccion.Opcode, instruccion.Parametros), log.INFO)
	
	//todo INSTRUCCIONES SYSCALLS
	if instruccion.Opcode == "IO" {
		Syscall_IO(instruccion)
	}
	if instruccion.Opcode == "INIT_PROC" {
		Syscall_Init_Proc(instruccion)
	}
	if instruccion.Opcode == "DUMP_MEMORY" {
		Syscall_Dump_Memory()
	}
	if instruccion.Opcode == "EXIT" {
		Syscall_Exit()
	}
	//todo OTRAS INSTRUCCIONES
	if instruccion.Opcode == "NOOP" {
	}

	if instruccion.Opcode == "GOTO" {
		pcNuevo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
			return
		}
		global.PCB_Actual.PC = pcNuevo
	}

	//todo INSTRUCCIONES MMU

	if (instruccion.Opcode == "READ") { /* READ 0 20 // READ (Dirección, Tamaño) */
		
		tlbHabilitada := global.CpuConfig.TlbEntries > 0
		cacheHabilitada := global.CpuConfig.CacheEntries > 0

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)

		if err != nil {
			fmt.Println("Error al convertir:", err)
		} else {
			ConfigMMU()
			desplazamiento = direccionLogica % configMMU.Tamanio_pagina
			nroPagina = direccionLogica / configMMU.Tamanio_pagina
			
		}
		if (cacheHabilitada)  {
			if(cacheHIT()){
				
			}
		} else if (tlbHabilitada) {
			if(tlbHIT()){
				direccionFisica = MMU(direccionLogica, instruccion.Opcode,nroPagina)
			}
		} else {
			tamanioStr := instruccion.Parametros[1]
			tamanio, err := strconv.Atoi(tamanioStr)
			
			direccionFisica = MMU(direccionLogica, instruccion.Opcode,nroPagina)
			
			if err != nil {
				fmt.Println("Error al convertir:", err)
			} else {
				MemoriaLee(direccionFisica, tamanio)
			}
		}
	}

	if (instruccion.Opcode == "WRITE") { /* WRITE 0 EJEMPLO_DE_ENUNCIADO // WRITE (Dirección, Datos) */
		
		tlbHabilitada := global.CpuConfig.TlbEntries > 0
		cacheHabilitada := global.CpuConfig.CacheEntries > 0

		direccionLogicaStr := instruccion.Parametros[0]
		datos := instruccion.Parametros[1]

		direccionLogica, err := strconv.Atoi(direccionLogicaStr)

		if err != nil {
			fmt.Println("Error al convertir:", err)
		} else {
			direccionFisica = MMU(direccionLogica, instruccion.Opcode,nroPagina)
		}
		if (cacheHabilitada)  {
			if(cacheHIT()){
				
			}
		} else if (tlbHabilitada) {
			if(tlbHIT() != -1){
				
			}
		} else {
			MemoriaEscribe(direccionFisica, datos)
		}
	}
}

/* func cacheHIT() bool {
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if(global.CACHE[i].NroPagina == nroPagina){
			return true
		}		
	}
	return false	
} */
 
func tlbHIT() int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if(global.TLB[i].NroPagina == nroPagina){
			return global.TLB[i].Marco
		}
	}
	return -1
}

func MemoriaLee(direccionFisica int, tamanio int) error {
	datosEnvio := estructuras.PedidoREAD{
		PID:             global.PCB_Actual.PID,
		DireccionFisica: direccionFisica,
		Tamanio:           tamanio,
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

func MMU(direccionLogica int, opcode string, nroPagina int) int {

		

		listaEntradas := armarListaEntradas(nroPagina)

		accederTabla := estructuras.AccesoTP{
			PID:      global.PCB_Actual.PID,
			Entradas: listaEntradas,
		}

		marco := pedirMarco(accederTabla)

		direccionFisica := marco*configMMU.Tamanio_pagina + desplazamiento

		return direccionFisica
}

func pedirMarco(estructuras.AccesoTP) int {
	var accesoTP estructuras.AccesoTP

	jsonData, err := json.Marshal(accesoTP)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return -1
	}

	url := fmt.Sprintf("http://%s:%d/pedirMarco", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                                //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: "+err.Error(), log.ERROR)
		return -1
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria de forma exitosa", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var marco int
	err = json.Unmarshal(body, &marco)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return -1
	}
	return marco
}

func armarListaEntradas(nroPagina int) []int {
	cantNiveles := configMMU.Cant_N_Niveles
	cantEntradas := configMMU.Cant_entradas_tabla

	entradas := make([]int, cantNiveles)

	for i := 1; i <= cantNiveles; i++ {
		entradas[i-1] = int(math.Floor(float64(nroPagina)/math.Pow(float64(cantEntradas), float64(cantNiveles-i)))) % cantEntradas
	}
	return entradas
}

func TLB(nroPagina int) {
	// conseguir el marco
	// ver si está la página
}

func ObtenerFrameDeMemoria(nroPagina int) {}

func ActualizarTLB(nroPagina int, marco int) {}

func CheckInterrupt() {}

func CacheDePaginas(direccionFisica int) {
	/* if(cacheHit){

	} else if (cacheMiss){
		AccederMemoria()
		ActualizarCache()
	} */
}

func AccederMemoria()  {}
func ActualizarCache() {}

var configMMU estructuras.ConfiguracionMMU

/*
type ConfiguracionMMU struct {
	Tamaño_página       int `json:"tamaño_página"`
	Cant_entradas_tabla int `json:"cant_entradas_tabla"`
	Cant_N_Niveles      int `json:"cant_N_Niveles"`
}
*/

func ConfigMMU() error {
	url := fmt.Sprintf("http://%s:%d/configuracionMMU", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)
	resp, err := http.Get(url)

	if err != nil {
		global.LoggerCpu.Log("Error al conectar con Memoria:", log.ERROR)
		return err
	}
	defer resp.Body.Close() //cierra automáticamente el cuerpo de la respuesta HTTP

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo respuesta de Memoria:", log.ERROR)
		return err
	}

	err = json.Unmarshal(body, &configMMU) // convierto el JSON que recibi de Memoria y lo guardo en el struct configMMU.
	if err != nil {
		global.LoggerCpu.Log("Error parseando JSON de configuración:", log.ERROR)
		return err
	}

	return nil
}

//todo DELEGO SYSCALLS

func Syscall_IO(instruccion Instruccion) {
	tiempo, err := strconv.Atoi(instruccion.Parametros[1])
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
		return
	}

	syscall_IO := estructuras.Syscall_IO{
		IoSolicitada:   instruccion.Parametros[0],
		TiempoEstimado: tiempo,
		PIDproceso:     global.PCB_Actual.PID,
	}

	jsonData, err := json.Marshal(syscall_IO)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/IO", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                      //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Init_Proc(instruccion Instruccion) {
	tamanio, err := strconv.Atoi(instruccion.Parametros[1]) //convieto tamanio de string a int
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tamanio: %v", log.ERROR)
		return
	}

	syscall_Init_Proc := estructuras.Syscall_Init_Proc{
		ArchivoInstrucciones: instruccion.Parametros[0],
		Tamanio:              tamanio,
		/* PIDproceso: pidActual, */
	}

	jsonData, err := json.Marshal(syscall_Init_Proc)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/Init_Proc", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                             //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Dump_Memory() {
	url := fmt.Sprintf("http://%s:%d/Dump_Memory?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel, global.PCB_Actual.PID) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil)                                                                                   //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Exit() {
	url := fmt.Sprintf("http://%s:%d/Exit?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel, global.PCB_Actual.PID) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil)                                                                            //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

/*
TODO:
solicitar a memoria utilizando solo el PC, query params
*/
