package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"strings"
	"time"
)

var tiempoInicio time.Time
var sumarPC bool = true

func CicloDeInstruccion() bool {
	var instruccionAEjecutar = Fetch()
	
	instruccion, requiereMMU := Decode(instruccionAEjecutar)

	tiempoInicio = time.Now()

	opcode,err := Execute(instruccion, requiereMMU)
	if err != nil {
		global.LoggerCpu.Log("Error ejecutando instrucción: "+err.Error(), log.ERROR)
		return false
	}
	if opcode == "EXIT"{
		global.LoggerCpu.Log("Es EXIT, corta el ciclo de instrucción", log.DEBUG)
		return false
	}

	CheckInterrupt()
	
	seguirEjecutando := instruccion.Opcode != "EXIT" && instruccion.Opcode != "IO" && instruccion.Opcode != "DUMP_MEMORY"
	return seguirEjecutando
}

func Fetch() string {
	pidActual := global.PCB_Actual.PID
	pcActual := global.PCB_Actual.PC

	global.LoggerCpu.Log(fmt.Sprintf("\033[36m## PID: %d - FETCH - Program Counter: %d\033[0m", pidActual, pcActual), log.INFO) //!! Fetch Instrucción - logObligatorio

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

func Execute(instruccion Instruccion, requiereMMU bool) (string, error) {

	global.LoggerCpu.Log(fmt.Sprintf("\033[36m## PID: %d - Ejecutando: %s - %s\033[0m", global.PCB_Actual.PID, instruccion.Opcode, instruccion.Parametros), log.INFO) //!! Instrucción Ejecutada - logObligatorio

	//todo INSTRUCCIONES SYSCALLS
	if instruccion.Opcode == "IO" {
		global.Motivo = "IO"
		global.Rafaga = float64(time.Since(tiempoInicio).Milliseconds())
		Desalojo()
		global.PCB_Actual.PC++
		sumarPC = false
		cortoProceso()
		Syscall_IO(instruccion)
		return "",nil
	}
	if instruccion.Opcode == "INIT_PROC" {
		Syscall_Init_Proc(instruccion)
		return "",nil
	}
	if instruccion.Opcode == "DUMP_MEMORY" {
		global.Motivo = "DUMP"
		global.Rafaga = float64(time.Since(tiempoInicio).Milliseconds())
		Desalojo()
		global.PCB_Actual.PC++
		sumarPC = false
		cortoProceso()
		Syscall_Dump_Memory()
		return "",nil
	}
	if instruccion.Opcode == "EXIT" {
		global.Motivo = "EXIT"
		global.Rafaga = float64(time.Since(tiempoInicio).Milliseconds())
		Desalojo()
		global.PCB_Actual.PC++
		sumarPC = false
		Syscall_Exit()
		DevolucionPID()
		global.LoggerCpu.Log(fmt.Sprintf("\033[35mProceso %d finalizado (EXIT). Fin del ciclo\033[0m",global.PCB_Actual.PID), log.INFO)
		return "EXIT",nil
	}

	//todo OTRAS INSTRUCCIONES
	if instruccion.Opcode == "NOOP" {
		return "",nil
	}

	if instruccion.Opcode == "GOTO" {
		pcNuevo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			return "",fmt.Errorf("error al convertir tiempo estimado")
		}
		global.PCB_Actual.PC = pcNuevo
		return "",nil
	}

	//todo INSTRUCCIONES MMU
	if requiereMMU {
		var desplazamiento int

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			return "",fmt.Errorf("error al convertir dirección logica")
		} else {
			if global.ConfigMMU.Tamanio_pagina == 0 {
				global.LoggerCpu.Log("Error: Tamanio_pagina es 0 antes de calcular el desplazamiento", log.ERROR)
				return "",nil
			}

			desplazamiento = direccionLogica % global.ConfigMMU.Tamanio_pagina
			nroPagina = direccionLogica / global.ConfigMMU.Tamanio_pagina
		}

		if instruccion.Opcode == "READ" { // READ 0 20 - READ (Dirección, Tamaño)
			READ(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada, direccionLogica)
		}

		if instruccion.Opcode == "WRITE" { // WRITE 0 EJEMPLO_DE_ENUNCIADO - WRITE (Dirección, Datos)
			WRITE(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada)
			global.LoggerCpu.Log(fmt.Sprintf("Contenido CACHE: %v", global.CACHE), log.DEBUG)
		}
	}
	return "",nil
}

func CheckInterrupt() {
	if global.Interrupcion {
		global.LoggerCpu.Log(("Hay interrupción"), log.DEBUG) 
		global.Motivo = "READY"
		global.Rafaga = float64(time.Since(tiempoInicio).Milliseconds())
		Desalojo()
		cortoProceso()
		global.PCB_Actual = global.PCB_Interrupcion
		global.Interrupcion = false
	}else{
		if(sumarPC){
			global.PCB_Actual.PC = global.PCB_Actual.PC + 1
		}
		global.LoggerCpu.Log(fmt.Sprintf("No hay interrupción, nuevo pc: %d", global.PCB_Actual.PC), log.DEBUG)
		sumarPC = true
	}
}