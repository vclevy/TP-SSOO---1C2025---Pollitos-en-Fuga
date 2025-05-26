package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

var instruccionesConMMU = map[string]bool{
	"WRITE":      true,
	"READ":       true,
}

var pidActual int
var pcActual int

func Fetch(pid int, pc int) {
	pidActual = pid
	pcActual = pc
		
	global.LoggerCpu.Log(fmt.Sprintf(" ## PID: %d - FETCH - Program Counter: %d", pidActual, pcActual), log.INFO)
	
	solicitudInstruccion := estructuras.SolicitudInstruccion{
		Pid: pid,
		Pc:  pc,
	}

	//petición
	jsonData, err := json.Marshal(solicitudInstruccion)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/solicitudInstruccion", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria de forma exitosa", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var instruccionAEjecutar string
	err = json.Unmarshal(body, &instruccionAEjecutar)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return
	}

	global.LoggerCpu.Log(fmt.Sprintf("Memoria respondió con la instrucción: %s", instruccionAEjecutar), log.INFO)

	Decode(instruccionAEjecutar)
}

type Instruccion struct {
	Opcode  string	`json:"opcode"`  // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción, de tipo variable
}

func Decode(instruccionAEjecutar string){
	instruccionPartida := strings.Fields(instruccionAEjecutar) //!!ver

	opcode := instruccionPartida[0]
	parametros := instruccionPartida[1:]

	instruccion := Instruccion{
		Opcode: opcode,
		Parametros:  parametros,
	}

	Execute(instruccion)
}

func Execute(instruccion Instruccion){
	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", pidActual, instruccion.Opcode , instruccion.Parametros), log.INFO)
  	
	//todo INSTRUCCIONES MMU
	/* 
	WRITE 0 EJEMPLO_DE_ENUNCIADO // WRITE (Dirección, Datos)
	READ 0 20 // READ (Dirección, Tamaño)
	*/

	if _, requiereMMU := instruccionesConMMU[instruccion.Opcode]; requiereMMU {
		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			fmt.Println("Error al convertir:", err)
		} else {
			MMU(direccionLogica)
		}
	}

  	//todo INSTRUCCIONES SYSCALLS
	if(instruccion.Opcode == "IO"){
		Syscall_IO(instruccion)
	}
	if(instruccion.Opcode == "INIT_PROC"){
		Syscall_Init_Proc(instruccion)
	}
	if(instruccion.Opcode == "DUMP_MEMORY"){
		Syscall_Dump_Memory()
	}
	if(instruccion.Opcode == "EXIT"){
		Syscall_Exit()
	}
	//todo OTRAS INSTRUCCIONES 
	if(instruccion.Opcode == "NOOP"){}
	
	if(instruccion.Opcode == "GOTO"){	
		pcNuevo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
			return
		}
		pcActual = pcNuevo
	}
}

func MMU(direccionLogica int){
	tlbHabilitada := global.CpuConfig.TlbEntries > 0
	tlbDeshabilitada := global.CpuConfig.TlbEntries == 0
	cacheHabilitada := global.CpuConfig.CacheEntries > 0
	cacheDeshabilitada := global.CpuConfig.CacheEntries == 0
	
	ConfigMMU()
	nroPagina := direccionLogica / configMMU.Tamaño_página
	desplazamiento := direccionLogica % configMMU.Tamaño_página
	var marco int
	if(tlbHabilitada){ //TLB habilitada
		marco = TLB()
		
	} else if (tlbDeshabilitada){
		marco = ObtenerFrameDeMemoria()
	} else{
		global.LoggerCpu.Log("Error de entradas TLB", log.ERROR)
		return
	}
	direccionFisica := marco * configMMU.Tamaño_página + desplazamiento

	if(cacheHabilitada){ //caché habilitada
		CacheDePaginas()
	}else if(cacheDeshabilitada){
		AccederMemoria()
	} else{
		global.LoggerCpu.Log("Error de entradas Cache", log.ERROR)
		return
	}
}

func CheckInterrupt(instruccion Instruccion){}

func TLB(){
	if(tlbHit){

	}else if (tlbMiss){
		/* marco := */ObtenerFrameDeMemoria()
		ActualizarTLB()
	}
}
func CacheDePaginas(){
	if(cacheHit){
				
	} else if (cacheMiss){
		AccederMemoria()
		ActualizarCache()
	}
}
func ObtenerFrameDeMemoria(){}
func AccederMemoria(){}
func ActualizarTLB(){}
func ActualizarCache(){}

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

func Syscall_IO(instruccion Instruccion){
	tiempo, err := strconv.Atoi(instruccion.Parametros[1])
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
		return
	}

	syscall_IO := estructuras.Syscall_IO{
		IoSolicitada : instruccion.Parametros[0],
		TiempoEstimado : tiempo,
		PIDproceso: pidActual,
	}

	jsonData, err := json.Marshal(syscall_IO)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/IO", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: " + err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Init_Proc(instruccion Instruccion){
	tamanio, err := strconv.Atoi(instruccion.Parametros[1]) //convieto tamanio de string a int
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tamanio: %v", log.ERROR)
		return
	}

	syscall_Init_Proc := estructuras.Syscall_Init_Proc{
		ArchivoInstrucciones : instruccion.Parametros[0],
		Tamanio : tamanio,
		/* PIDproceso: pidActual, */
	}

	jsonData, err := json.Marshal(syscall_Init_Proc)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/Init_Proc", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: " + err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Dump_Memory(){
	url := fmt.Sprintf("http://%s:%d/Dump_Memory?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel,pidActual) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: " + err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Exit(){
	url := fmt.Sprintf("http://%s:%d/Exit?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel,pidActual) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: " + err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

/* 
TODO:
? implementar que las funciones reciban errores(?) func Decode(instruccion string) (string, error) 
? que es lo que hace arrancar el fetch? Por ahora es el handshake con kernel
*/