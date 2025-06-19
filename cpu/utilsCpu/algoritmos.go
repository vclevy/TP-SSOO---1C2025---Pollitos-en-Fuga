package utilsIo

import (
	"github.com/sisoputnfrba/tp-golang/cpu/global"
)

var fifoIndice int = 0
var indiceLRU int = 0
var punteroClock int = 0
var punteroClockModificado int = 0

func AlgoritmoTLB() int { // la página no está en la tlb
	if indiceVacio() == -1 { // no hay indice vacio
		if global.CpuConfig.TlbReplacement == "FIFO" {

			indice := fifoIndice

			fifoIndice = (fifoIndice + 1) % len(global.TLB)

			return indice

		} else if global.CpuConfig.TlbReplacement == "LRU" {
			minUso := global.TLB[0].UltimoUso
			for i := 1; i < len(global.TLB); i++ {
				if global.TLB[i].UltimoUso < minUso {
					minUso = global.TLB[i].UltimoUso
					indiceLRU = i
				}
			}
			
			lruCounter++
			global.TLB[indiceLRU].UltimoUso = lruCounter
			return indiceLRU
		}
	}
	return indiceVacio()
}

func AlgoritmoCACHE() int { //CACHE: CLOCK o CLOCK-M
	if indiceVacio() == -1 {
    if(global.CpuConfig.CacheReplacement == "CLOCK") {
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
    } else if(global.CpuConfig.CacheReplacement == "CLOCK-M"){
		// Paso 1: Buscar (U=0, M=0) sin modificar ningún bit
			for i := 0; i < len(global.CACHE); i++ {
				pos := (punteroClockModificado + i) % len(global.CACHE)
				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 0 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
					return pos
				}
			}

			// Paso 2: Buscar (U=0, M=1) y poner U=0 mientras se recorre
			for i := 0; i < len(global.CACHE); i++ {
				pos := (punteroClockModificado + i) % len(global.CACHE)
				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 1 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
					return pos
				}
				// Mientras recorro, pongo BitUso en 0
				global.CACHE[pos].BitUso = 0
			}

			// Paso 3: Volver al paso 1
			for i := 0; i < len(global.CACHE); i++ {
				pos := (punteroClockModificado + i) % len(global.CACHE)
				if global.CACHE[pos].BitUso == 0 && global.CACHE[pos].BitModificado == 0 {
					punteroClockModificado = (pos + 1) % len(global.CACHE)
					return pos
				}
			}
		
	}	
	}
	return indiceVacio()
}

func indiceVacio() int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.CACHE[i].NroPagina == -1 {
			return i
		}
	}
	return -1
}
