package utilsIo

import (
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var fifoIndice int = 0
var punteroClock int = 0
var punteroClockModificado int = 0
var lruCounter int = 0

func AlgoritmoTLB() int { // la página no está en la tlb y no hay indice vacio
	if global.CpuConfig.TlbReplacement == "FIFO" {
		indice := fifoIndice
		fifoIndice = (fifoIndice + 1) % len(global.TLB)
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
		return indiceLRU
	} else {
		global.LoggerCpu.Log("El algoritmo no es FIFO ni LRU", log.ERROR)
		return -1
	}
}

func AlgoritmoCACHE() int { //CACHE: CLOCK o CLOCK-M
	if global.CpuConfig.CacheReplacement == "CLOCK" {
		for {
			if global.CACHE[punteroClock].BitUso == 0 {
				indice := punteroClock
				punteroClock = (punteroClock + 1) % len(global.CACHE)
				return indice
			} else {
				global.CACHE[punteroClock].BitUso = 0
				punteroClock = (punteroClock + 1) % len(global.CACHE)
			}
		}
	} else if global.CpuConfig.CacheReplacement == "CLOCK-M" {
		for {
			for i := 0; i < len(global.CACHE); i++ { // Busco página con U=0, M=0
				pos := (punteroClockModificado + i) % len(global.CACHE)
				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 0 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
					return pos
				}
			}
			for i := 0; i < len(global.CACHE); i++ { // Busco página con U=0, M=1 y poner U=0 mientras se recorre
				pos := (punteroClockModificado + i) % len(global.CACHE)

				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 1 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
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
