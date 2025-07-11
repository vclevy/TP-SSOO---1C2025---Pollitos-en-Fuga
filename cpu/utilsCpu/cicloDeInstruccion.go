package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"strings"
)

var sumarPC bool = true

func CicloDeInstruccion() bool {

	var instruccionAEjecutar = Fetch()

	instruccion, requiereMMU := Decode(instruccionAEjecutar)

	opcode, err := Execute(instruccion, requiereMMU)
	if err != nil {
		global.LoggerCpu.Log("Error ejecutando instrucción: "+err.Error(), log.ERROR)
		return false
	}
	if opcode == "EXIT" {
		global.LoggerCpu.Log("Es EXIT, corta el ciclo de instrucción", log.INFO)
		return false
	}

	CheckInterrupt()

	seguirEjecutando := instruccion.Opcode != "EXIT" && instruccion.Opcode != "IO" && instruccion.Opcode != "DUMP_MEMORY"
	return seguirEjecutando
}

func Fetch() string {
	pidActual := global.PCB_Actual.PID
	pcActual := global.PCB_Actual.PC

	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - FETCH - Program Counter: %d", pidActual, pcActual), log.INFO) //!! Fetch Instrucción - logObligatorio

	solicitudInstruccion := estructuras.PCB{
		PID: pidActual,
		PC:  pcActual,
	}

	var instruccionAEjecutar = instruccionAEjecutar(solicitudInstruccion)

	return instruccionAEjecutar
}

func Decode(instruccionAEjecutar string) (Instruccion, bool) {
	instruccionPartida := strings.Fields(instruccionAEjecutar) //?  "MOV AX BX" --> []string{"MOV", "AX", "BX"}

	if len(instruccionPartida) == 0 {
		return Instruccion{}, false
	}

	instruccion := Instruccion{
		Opcode:     instruccionPartida[0],
		Parametros: instruccionPartida[1:],
	}

	_, requiereMMU := instruccionesConMMU[instruccion.Opcode]

	return instruccion, requiereMMU
}

func Execute(instruccion Instruccion, requiereMMU bool) (string, error) {

	global.LoggerCpu.Log(fmt.Sprintf("## PID: %d - Ejecutando: %s - %s", global.PCB_Actual.PID, instruccion.Opcode, instruccion.Parametros), log.INFO) //!! Instrucción Ejecutada - logObligatorio
	if global.PCB_Actual == nil {
		return "", fmt.Errorf("PCB_Actual es nil: no se puede ejecutar instrucción")
	}
	if instruccion.Opcode == "IO" {
		sumarPC = false
		global.Motivo = "IO"
		global.PCB_Actual.PC++

		tiempo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
			return "", err
		}

		global.IO_Request = estructuras.Syscall_IO{
			PIDproceso:     global.PCB_Actual.PID,
			IoSolicitada:   instruccion.Parametros[0],
			TiempoEstimado: tiempo,
		}

		cortoProceso()
		Desalojo()
		return "", nil
	}

	if instruccion.Opcode == "INIT_PROC" {
		sumarPC = true
		Syscall_Init_Proc(instruccion)
		//cortoProceso()

		return "", nil
	}
	if instruccion.Opcode == "DUMP_MEMORY" {
		sumarPC = false
		global.Motivo = "DUMP"
		global.PCB_Actual.PC++
		cortoProceso()
		Desalojo()

		return "", nil
	}

	if instruccion.Opcode == "EXIT" {
		sumarPC = false
		global.Motivo = "EXIT"
		pid := global.PCB_Actual.PID // antes de que se borre

		cortoProceso()
		Desalojo()

		global.LoggerCpu.Log(fmt.Sprintf("Proceso %d finalizado (EXIT). Fin del ciclo", pid), log.INFO)

		return "EXIT", nil
	}

	if instruccion.Opcode == "NOOP" {
		sumarPC = true
		return "", nil
	}

	if instruccion.Opcode == "GOTO" {
		sumarPC = false
		if len(instruccion.Parametros) < 1 {
			return "", fmt.Errorf("GOTO requiere 1 parámetro, recibido: %v", instruccion.Parametros)
		}

		pcNuevo, err := strconv.Atoi(instruccion.Parametros[0])
		if err != nil {
			return "", fmt.Errorf("error al convertir tiempo estimado")
		}
		global.PCB_Actual.PC = pcNuevo
		return "", nil
	}

	if requiereMMU {
		sumarPC = true
		var desplazamiento int

		if len(instruccion.Parametros) < 1 {
			return "", fmt.Errorf("instrucción requiere al menos 1 parámetro, recibido: %v", instruccion.Parametros)
		}

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			return "", fmt.Errorf("error al convertir dirección logica")
		} else {
			if global.ConfigMMU.Tamanio_pagina == 0 {
				global.LoggerCpu.Log("Error: Tamanio_pagina es 0 antes de calcular el desplazamiento", log.ERROR)
				return "", nil
			}

			desplazamiento = direccionLogica % global.ConfigMMU.Tamanio_pagina
			nroPagina = direccionLogica / global.ConfigMMU.Tamanio_pagina
		}

		if instruccion.Opcode == "READ" { // READ 0 20 - READ (Dirección, Tamaño)
			READ(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada, direccionLogica)
		}

		if instruccion.Opcode == "WRITE" { // WRITE 0 EJEMPLO_DE_ENUNCIADO - WRITE (Dirección, Datos)
			WRITE(instruccion, global.CacheHabilitada, desplazamiento, global.TlbHabilitada)
			/* global.LoggerCpu.Log(fmt.Sprintf("Contenido CACHE: %v", global.CACHE), log.DEBUG)*/
		}
	}
	return "", nil
}

func CheckInterrupt() {
	if global.Interrupcion {
		global.Motivo = "READY"
		Desalojo()
		cortoProceso()
		global.PCB_Actual = global.PCB_Interrupcion
		global.Interrupcion = false
	} else {
		if sumarPC {
			global.PCB_Actual.PC = global.PCB_Actual.PC + 1
		}
	}
}
