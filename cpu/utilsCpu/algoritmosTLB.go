package utilsIo

import (
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
)


func AlgoritmoTLB(tlb []estructuras.DatoTLB){ //TLB: FIFO o LRU.
	if(global.CpuConfig.TlbReplacement == "FIFO"){

	}else if(global.CpuConfig.TlbReplacement == "LRU"){
		
	}else{
		
	}
}

func AlgoritmoCACHE(cache []estructuras.DatoCACHE){ //CACHE: CLOCK o CLOCK-M
	if(global.CpuConfig.CacheReplacement == "CLOCK"){

	}else if(global.CpuConfig.CacheReplacement == "CLOCK-M"){
		
	}else{
		
	}
}