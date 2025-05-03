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
	/* 	"math" */)

var instruccionesConMMU = map[string]bool{
	"WRITE":      true,
	"READ":       true,
}

var instruccionesSyscall = map[string]bool{
	"IO": true,
	"INIT_PROC": true,
	"DUMP_MEMORY": true,
	"EXIT": true,
}

var pidEnEjecucion int

func Fetch(pid int, pc int) {
	
	global.LoggerCpu.Log(fmt.Sprintf(" ## PID: %d - FETCH - Program Counter: %d", pid, pc), log.INFO)
	
	solicitudInstruccion := estructuras.Instruccion{
		Pid: pid,
		Pc:  pc,
	}

	pidEnEjecucion = pid

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
	instruccionPartida := strings.Fields(instruccionAEjecutar) //ver

	opcode := instruccionPartida[0]
	parametros := instruccionPartida[1:]

	instruccion := Instruccion{
		Opcode: opcode,
		Parametros:  parametros,
	}

	Execute(instruccion)
}

func Execute(instruccion Instruccion){
	if _, requiereMMU := instruccionesConMMU[instruccion.Opcode]; requiereMMU {
		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			fmt.Println("Error al convertir:", err)
		} else {
			MMU(direccionLogica)
		}
	}
	
}

func MMU(direccionLogica int){
	/* nro_pagina := math.Floor(float64(direccionLogica) / float64(configMMU.Tamaño_página)) 
	desplazamiento := direccionLogica % configMMU.Tamaño_página */
	/* traducir direcciones lógicas a físicas, 
		dirección logica [entrada_nivel_1 | entrada_nivel_2 | … | entrada_nivel_X | desplazamiento] 
		
		Teniendo una cantidad de niveles N y un identificador X de cada nivel podemos utilizar las siguientes fórmulas:
		nro_página = floor(dirección_lógica / tamaño_página)
		entrada_nivel_X = floor(nro_página  / cant_entradas_tabla ^ (N - X)) % cant_entradas_tabla
		desplazamiento = dirección_lógica % tamaño_página
	*/
/* 	return 0 */
}

func CheckInterrupt(instruccion Instruccion){}

func EnviarInstruccionAKernel(instruccion Instruccion,  pid int){
	type Syscall struct {
		/* IDCpu  	 */	
		Instruccion	 Instruccion `json:"Instruccion"`
		Pid		int		`json:"Pid"`
	}
	syscall := Syscall{
		/* IDCpu  */
		Pid: pid,
		Instruccion:  instruccion,
	}

	jsonData, err := json.Marshal(syscall)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/envioSyscall", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: " + err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

type configuracionMMU struct {
	Tamaño_página 			int 	`json:"tamaño_página"`
	Cant_entradas_tabla  	int     `json:"cant_entradas_tabla"`
	Cant_N_Niveles    		int     `json:"cant_N_Niveles"`
}
var configMMU configuracionMMU

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

/* 
TODO:
? a kernel le paso el struct o el string?
? usar query paths
? implementar que las funciones reciban errores(?) func Decode(instruccion string) (string, error) 
? hacer mmu
? delegar las syscalls a kernel, me devuelve algo kernel?
*/