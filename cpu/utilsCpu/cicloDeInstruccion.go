package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"io"
	"net/http"
)

var tiempoInicio time.Time
var ConfigMMU estructuras.ConfiguracionMMU

func CicloDeInstruccion() bool {
	global.LoggerCpu.Log(("🟨Comienza ciclo instruccion"), log.INFO)

	var instruccionAEjecutar = Fetch()
	
	instruccion, requiereMMU := Decode(instruccionAEjecutar)

	tiempoInicio = time.Now()

	/* if(instruccion.Opcode == "EXIT"){		
		err := Execute(instruccion, requiereMMU)
		if err != nil {
			global.LoggerCpu.Log("Error ejecutando instrucción: "+err.Error(), log.ERROR)
			return false
		}
		CheckInterrupt()
		global.LoggerCpu.Log("🟦 Proceso finalizado (EXIT). Fin del ciclo", log.INFO)
	} */
	
	err := Execute(instruccion, requiereMMU)
	if err != nil {
		global.LoggerCpu.Log("Error ejecutando instrucción: "+err.Error(), log.ERROR)
		return false
	}

	CheckInterrupt()
	global.LoggerCpu.Log("🟧Termina ciclo instruccion", log.INFO)
	
	return instruccion.Opcode != "EXIT"
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
		return nil
	}
	if instruccion.Opcode == "INIT_PROC" {
		Syscall_Init_Proc(instruccion)
		return nil
	}
	if instruccion.Opcode == "DUMP_MEMORY" {
		Syscall_Dump_Memory()
		return nil
	}
	if instruccion.Opcode == "EXIT" {
		global.Motivo = "EXIT"
		global.Rafaga = time.Since(tiempoInicio).Seconds()		
		Syscall_Exit()
		DevolucionPID()
		global.LoggerCpu.Log("🟦 Proceso finalizado (EXIT). Fin del ciclo", log.INFO)
		return nil
	}

	//todo OTRAS INSTRUCCIONES
	if instruccion.Opcode == "NOOP" {
		return nil
	}

	if instruccion.Opcode == "GOTO" {
		pcNuevo, err := strconv.Atoi(instruccion.Parametros[1])
		if err != nil {
			return fmt.Errorf("error al convertir tiempo estimado")
		}
		global.PCB_Actual.PC = pcNuevo
		return nil
	}

	//todo INSTRUCCIONES MMU
	if requiereMMU {
		var desplazamiento int

		direccionLogicaStr := instruccion.Parametros[0]
		direccionLogica, err := strconv.Atoi(direccionLogicaStr)
		if err != nil {
			return fmt.Errorf("error al convertir dirección logica")
		} else {
			/* err := ConfigMMU()
			if err != nil {
				    global.LoggerCpu.Log("Error en ConfigMMU: "+err.Error(), log.ERROR)

			} */

			err := CargarConfigMMU()
			if err != nil {
				    global.LoggerCpu.Log("Error en ConfigMMU: "+err.Error(), log.ERROR)

			}

			if ConfigMMU.Tamanio_pagina == 0 {
				global.LoggerCpu.Log("Error: Tamanio_pagina es 0 antes de calcular el desplazamiento", log.ERROR)
				return nil
			}

			desplazamiento = direccionLogica % ConfigMMU.Tamanio_pagina
			nroPagina = direccionLogica / ConfigMMU.Tamanio_pagina
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
		global.LoggerCpu.Log(("Hay interrupción"), log.INFO) 
		global.Motivo = "READY"
		global.Rafaga = time.Since(tiempoInicio).Seconds()
		cortoProceso()
		global.PCB_Actual = global.PCB_Interrupcion
		global.Interrupcion = false
	}else{
		global.PCB_Actual.PC = global.PCB_Actual.PC + 1
	}
}

func CargarConfigMMU() error {
	url := fmt.Sprintf("http://%s:%d/configuracionMMU", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		global.LoggerCpu.Log("Error al conectar con Memoria:", log.ERROR)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo respuesta de Memoria:", log.ERROR)
		return err
	}

	global.LoggerCpu.Log("JSON recibido de Memoria: "+string(body), log.DEBUG)

	err = json.Unmarshal(body, &ConfigMMU)
	if err != nil {
		global.LoggerCpu.Log("Error parseando JSON de configuracion: "+err.Error(), log.ERROR)
		return err
	}
	global.LoggerCpu.Log(fmt.Sprintf("Entradas tabla %d", ConfigMMU.Cant_entradas_tabla), log.DEBUG)
	global.LoggerCpu.Log(fmt.Sprintf("tamanio pagina %d", ConfigMMU.Tamanio_pagina), log.DEBUG)
	global.LoggerCpu.Log(fmt.Sprintf("cantidad niveles %d", ConfigMMU.Cant_N_Niveles), log.DEBUG)

	global.TamPagina = ConfigMMU.Tamanio_pagina
	return nil
}