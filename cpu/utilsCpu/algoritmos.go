package utilsIo

import (
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
/* 	"fmt" */
)

var fifoIndice int = 0

func AlgoritmoTLB() int { //TLB: FIFO o LRU.
	if(global.CpuConfig.TlbReplacement == "FIFO"){
		
		indice := fifoIndice
		
		fifoIndice = (fifoIndice + 1) % len(global.TLB)

/* 		fmt.Sprintf("Reemplazado el marco en TLB[ %d ]", indice)
 */		
		return indice
		
	}else if(global.CpuConfig.TlbReplacement == "LRU"){
		
	}else{
		
	}
	return 0
}

func AlgoritmoCACHE() int{ //CACHE: CLOCK o CLOCK-M
	/* if(global.CpuConfig.CacheReplacement == "CLOCK"){

	}else if(global.CpuConfig.CacheReplacement == "CLOCK-M"){
		
	}else{
		
	} */
	 return 0
}

func HayPaginaVacia(tlb []estructuras.DatoTLB){

}