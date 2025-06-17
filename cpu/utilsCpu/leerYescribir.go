package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

func WRITE(instruccion Instruccion, cacheHabilitada bool, desplazamiento int, tlbHabilitada bool) {
	dato := instruccion.Parametros[1]
	if cacheHabilitada {
		if CacheHIT(nroPagina) {
			indicePaginaEnCache(nroPagina)
			global.CACHE[indice].Contenido = dato
			global.CACHE[indice].BitModificado = 1


		} else {
			actualizarCACHE(nroPagina, dato)
		}
	} else {
		if tlbHabilitada {
			var marco int
			if TlbHIT(nroPagina) {
				marco = global.TLB[indice].Marco
			} else {
				marco = CalcularMarco()
			}
			direccionFisica = MMU(desplazamiento, marco)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, dato), log.INFO) //!! CACHE MISS

			MemoriaEscribe(direccionFisica, dato)
			actualizarTLB(nroPagina, marco)
		} else {
			marco := CalcularMarco()
			direccionFisica = MMU(desplazamiento, marco)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, dato), log.INFO) //!! CACHE MISS
			MemoriaEscribe(direccionFisica, dato)
		}
	}
}

func READ(instruccion Instruccion, cacheHabilitada bool, desplazamiento int, tlbHabilitada bool, direccionLogica int) {
	var marco int
	tamanioStr := instruccion.Parametros[1]
	tamanio, err := strconv.Atoi(tamanioStr)
	if err != nil {
		global.LoggerCpu.Log("error al convertir tamanio", log.ERROR)
	}

	if cacheHabilitada {
		if CacheHIT(nroPagina) {
			indice := indicePaginaEnCache(nroPagina)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, global.CACHE[indice].Contenido), log.INFO) //!! CACHE MISS
		} else {
			if tlbHabilitada {
				if TlbHIT(nroPagina) {
					marco = global.TLB[indice].Marco
					direccionFisica = MMU(desplazamiento, marco)
					MemoriaLee(direccionFisica, tamanio)
					actualizarTLB(nroPagina, marco)
					actualizarCACHE(nroPagina, global.CACHE[indice].Contenido)
				} else {
					marco = CalcularMarco()
					direccionFisica = marco * configMMU.Tamanio_pagina
					MemoriaLee(direccionFisica, tamanio)
					actualizarTLB(nroPagina, marco)
					actualizarCACHE(nroPagina, global.CACHE[indice].Contenido)
				}
			} else {
				marco = CalcularMarco()
				direccionFisica = marco * configMMU.Tamanio_pagina
				MemoriaLee(direccionFisica, tamanio)
				actualizarCACHE(nroPagina, global.CACHE[indice].Contenido)
			}
		}
	} else {
		if tlbHabilitada {
			if TlbHIT(nroPagina) {
				marco = CalcularMarco()
				direccionFisica = marco * configMMU.Tamanio_pagina
				MemoriaLee(direccionFisica, tamanio)
				actualizarTLB(nroPagina, marco)
			} else {
				marco = CalcularMarco()
				direccionFisica = marco * configMMU.Tamanio_pagina
				MemoriaLee(direccionFisica, tamanio)
				actualizarTLB(nroPagina, marco)
			}
		} else {
			marco = CalcularMarco()
			direccionFisica = marco * configMMU.Tamanio_pagina
			MemoriaLee(direccionFisica, tamanio)
		}
	}
}
