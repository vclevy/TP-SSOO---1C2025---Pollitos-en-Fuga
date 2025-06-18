package utilsIo

import (
	/* "github.com/sisoputnfrba/tp-golang/utils/estructuras" */
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	/* 	"fmt" */)

var fifoIndice int = 0
var indiceLRU int = 0
var punteroClock int = 0

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
		return 0
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
