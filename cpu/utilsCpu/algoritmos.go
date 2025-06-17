package utilsIo

import (
	/* "github.com/sisoputnfrba/tp-golang/utils/estructuras" */
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	/* 	"fmt" */)

var fifoIndice int = 0

func AlgoritmoTLB() int { // la página no está en la tlb
	if indiceVacio() == -1 { // no hay indice vacio
		if global.CpuConfig.TlbReplacement == "FIFO" {

			indice := fifoIndice

			fifoIndice = (fifoIndice + 1) % len(global.TLB)

			return indice

		} else if global.CpuConfig.TlbReplacement == "LRU" {
			indiceLRU := 0
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
	/* if(global.CpuConfig.CacheReplacement == "CLOCK"){

	}else if(global.CpuConfig.CacheReplacement == "CLOCK-M"){

	} */
	return 0
}

func indiceVacio() int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.CACHE[i].NroPagina == -1 {
			return i
		}
	}
	return -1
}
