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

	global.LoggerCpu.Log(fmt.Sprintf(" ## PID: %d - FETCH - Program Counter: %d", pidActual, pcActual), log.INFO) //!! Fetch Instrucción - logObligatorio

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

	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", global.PCB_Actual.PID, instruccion.Opcode, instruccion.Parametros), log.INFO) //!! Instrucción Ejecutada - logObligatorio

	//todo INSTRUCCIONES SYSCALLS
	if instruccion.Opcode == "IO" {
		global.Motivo = "BLOCKED"
		global.Rafaga = time.Since(tiempoInicio).Seconds()
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
		global.Rafaga = time.Since(tiempoInicio).Seconds()		
		DevolucionPID()
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
		var desplazamiento int

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			return fmt.Errorf("error al convertir dirección logica")
		} else {
			err := ConfigMMU()
			if err != nil {
				    global.LoggerCpu.Log("Error en ConfigMMU: "+err.Error(), log.ERROR)

			}
			desplazamiento = direccionLogica % configMMU.Tamanio_pagina
			nroPagina = direccionLogica / configMMU.Tamanio_pagina
		}

		if instruccion.Opcode == "READ" { // READ 0 20 - READ (Dirección, Tamaño)
			READ(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada, direccionLogica)
		}

		if instruccion.Opcode == "WRITE" { // WRITE 0 EJEMPLO_DE_ENUNCIADO - WRITE (Dirección, Datos)
			WRITE(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada)
		}
	}
	return nil
}

func CheckInterrupt() {
	if global.Interrupcion {
		global.Motivo = "READY"
		global.Rafaga = time.Since(tiempoInicio).Seconds()
		cortoProceso()
		global.PCB_Actual = global.PCB_Interrupcion
		global.Interrupcion = false
	}
}