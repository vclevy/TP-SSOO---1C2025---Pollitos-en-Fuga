package utilsIo

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var fifoIndice int = 0
var punteroClock int = 0
var punteroClockModificado int = 0
var lruCounter int = 0

func AlgoritmoTLB() int { // la página no está en la tlb y no hay indice vacio

	/* global.LoggerCpu.Log(fmt.Sprintf("Contenido TLB ANTES DE REEMPLAZAR: %v", global.TLB), log.DEBUG) */

	if global.CpuConfig.TlbReplacement == "FIFO" {
		indice := fifoIndice
		fifoIndice = (fifoIndice + 1) % len(global.TLB)
		global.LoggerCpu.Log(fmt.Sprintf("INDICE A REEMPLAZAR EN TLB POR ALGORTIMO %s es %d", global.CpuConfig.TlbReplacement, indice), log.ERROR)
		return indice
	} else if global.CpuConfig.TlbReplacement == "LRU" {
		var indiceLRU int = 0
		minUso := global.TLB[0].UltimoUso
		for i := 1; i < len(global.TLB); i++ { // me fijo que indice fue el ultimo en usarse
			if global.TLB[i].UltimoUso < minUso {
				minUso = global.TLB[i].UltimoUso
				indiceLRU = i
			}
		}
		global.LoggerCpu.Log(fmt.Sprintf("Ultimo uso del indice %d es %d", indiceLRU, global.TLB[indiceLRU].UltimoUso), log.ERROR)
		global.LoggerCpu.Log(fmt.Sprintf("INDICE A REEMPLAZAR EN TLB POR ALGORTIMO %s es %d", global.CpuConfig.TlbReplacement, indiceLRU), log.ERROR)
		return indiceLRU
	} else {
		global.LoggerCpu.Log("El algoritmo no es FIFO ni LRU", log.ERROR)
		return -1
	}
}

func AlgoritmoCACHE() int { //CACHE: CLOCK o CLOCK-M
	/* global.LoggerCpu.Log(fmt.Sprintf("Contenido CACHE ANTES DE REEMPLAZAR: %v", global.CACHE), log.DEBUG)
	 */
	if global.CpuConfig.CacheReplacement == "CLOCK" {
		for {
			global.LoggerCpu.Log(fmt.Sprintf("Bit de uso de la página %d: %d", global.CACHE[punteroClock].NroPagina, global.CACHE[punteroClock].BitUso), log.ERROR)

			if global.CACHE[punteroClock].BitUso == 0 { //BIT DE USO 0
				indice := punteroClock
				punteroClock = (punteroClock + 1) % len(global.CACHE)
				global.LoggerCpu.Log(fmt.Sprintf("INDICE A REEMPLAZAR EN CACHE POR ALGORTIMO %s es %d", global.CpuConfig.CacheReplacement, indice), log.ERROR)

				return indice
			} else { //BIT DE USO 1
				global.CACHE[punteroClock].BitUso = 0
				punteroClock = (punteroClock + 1) % len(global.CACHE)
			}
		}
	} else if global.CpuConfig.CacheReplacement == "CLOCK-M" {
		for {
			for i := 0; i < len(global.CACHE); i++ { // Busco página con U=0, M=0
				pos := (punteroClockModificado + i) % len(global.CACHE)
				global.LoggerCpu.Log(fmt.Sprintf("Bit de uso de la página %d: %d", global.CACHE[pos].NroPagina, global.CACHE[pos].BitUso), log.ERROR)
				global.LoggerCpu.Log(fmt.Sprintf("Bit modificado de la página %d: %d", global.CACHE[pos].NroPagina, global.CACHE[pos].BitModificado), log.ERROR)
				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 0 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
					global.LoggerCpu.Log(fmt.Sprintf("INDICE A REEMPLAZAR EN CACHE POR ALGORTIMO %s es %d", global.CpuConfig.CacheReplacement, pos), log.ERROR)

					return pos
				}
			}
			for i := 0; i < len(global.CACHE); i++ { // Busco página con U=0, M=1 y poner U=0 mientras se recorre
				pos := (punteroClockModificado + i) % len(global.CACHE)

				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 1 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)

					global.LoggerCpu.Log(fmt.Sprintf("INDICE A REEMPLAZAR EN CACHE POR ALGORTIMO %s es %d", global.CpuConfig.CacheReplacement, pos), log.ERROR)
					return pos
				}
				global.CACHE[pos].BitUso = 0 // Pongo BitUso en 0 mientras se recorre
			}
		}
	} else {
		global.LoggerCpu.Log("El algoritmo no es CLOCK ni CLOCK-M", log.ERROR)
		return -1
	}
}
