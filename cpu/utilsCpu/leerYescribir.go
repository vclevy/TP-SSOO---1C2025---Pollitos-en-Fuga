package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"time"
)

func WRITE(instruccion Instruccion, cacheHabilitada bool, desplazamiento int, tlbHabilitada bool) {
	datos := instruccion.Parametros[1]
	//bit uso ver
	if cacheHabilitada { //CACHE HABILITADA
		if CacheHIT(nroPagina) {//CACHE HIT
			indice := indicePaginaEnCache(nroPagina)			
			escribirCache(indice, datos, desplazamiento)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, 0, datos), log.INFO) //!! Lectura/Escritura Memoria - logObligatorio	
			
		} else {//CACHE MISS
			indiceEscribir,dirFisicaSinDespl := actualizarCACHE()
			escribirCache(indiceEscribir, datos, desplazamiento)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: ESCRIBIR - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, (dirFisicaSinDespl + desplazamiento), datos), log.INFO)
		}
	}else { //CACHE DESHABILITADA
		marco := CalcularMarco()
		direccionFisica := MMU(desplazamiento, marco)
		MemoriaEscribe(direccionFisica, datos)
	}
}

func READ(instruccion Instruccion, cacheHabilitada bool, desplazamiento int, tlbHabilitada bool, direccionLogica int) {
	tamanioStr := instruccion.Parametros[1]
    tamanio, err := strconv.Atoi(tamanioStr)

	if err != nil {
        global.LoggerCpu.Log("error al convertir tamanio", log.ERROR)
    }
	if(global.CacheHabilitada){//CACHE HABILITADA
		if CacheHIT(nroPagina) {//CACHE HIT
			indice := indicePaginaEnCache(nroPagina)
				paginaCompleta := global.CACHE[indice].Contenido
				if desplazamiento+tamanio > len(paginaCompleta) {
					global.LoggerCpu.Log("❌ Lectura fuera de rango en caché", log.ERROR)
					return
				}
				lectura := paginaCompleta[desplazamiento : desplazamiento + tamanio]
				global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, 0, lectura), log.INFO) //!! LECTURA SIN ACCEDER A MEMORIA (Desde caché)
		} else {//CACHE MISS
			indice, dirFisicaSinDespl := actualizarCACHE()
			paginaCompleta := global.CACHE[indice].Contenido
			lectura := paginaCompleta[desplazamiento : desplazamiento + tamanio]
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, dirFisicaSinDespl + desplazamiento, lectura), log.INFO)
		}
	}else { //CACHE DESHABILITADA
		marco := CalcularMarco()
		direccionFisica := MMU(desplazamiento, marco)
		MemoriaLee(direccionFisica, tamanio)
	}
}

func TlbHIT(pagina int) bool {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //!! TLB Hit - logObligatorio
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //!! TLB Miss - logObligatorio
	return false
}

func CacheHIT(pagina int) bool {
	time.Sleep(time.Millisecond * time.Duration(global.CpuConfig.CacheDelay)) 
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //!! Página encontrada en Caché - logObligatorio (Cache hit)
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //!! Página faltante en Caché - logObligatorio (Cache miss)
	return false
}

func actualizarCACHE() (int,int){ //
	time.Sleep(time.Millisecond * time.Duration(global.CpuConfig.CacheDelay))
	indicePisar := indiceVacioCACHE() 
	
	if indicePisar == -1 { // no hay espacio vacio en cachce
		indicePisar = AlgoritmoCACHE()
	}
	if global.CACHE[indicePisar].BitModificado == 1 {
		desalojar(indicePisar)
	}
	
	marco := CalcularMarco()
	dirFisicaSinDesplazamiento := MMU(0,marco)
	lecturaPagina := MemoriaLeePaginaCompleta(dirFisicaSinDesplazamiento)
	global.LoggerCpu.Log(fmt.Sprintf("pagina completa que se lee de memoria: %d", lecturaPagina), log.INFO)

	global.CACHE[indicePisar].NroPagina = nroPagina
	global.CACHE[indicePisar].Contenido = lecturaPagina
	global.CACHE[indicePisar].BitModificado = 0

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", global.PCB_Actual.PID, nroPagina), log.INFO) //!! Página ingresada en Caché - logObligatorio

	return indicePisar, dirFisicaSinDesplazamiento
}
//PISA UN VALOR DE TLB Y SE LO TRAE
func actualizarTLB() int{	
	indicePisar := indiceVacioTLB()
	/* lruCounter ++ */
	if indicePisar == -1 {
		indicePisar = AlgoritmoTLB()
	}
	lruCounter++
	marco := BuscarMarcoEnMemoria(nroPagina)
	global.TLB[indicePisar].Marco = marco
	global.TLB[indicePisar].NroPagina = nroPagina
	global.TLB[indicePisar].UltimoUso = lruCounter
	return marco
}


func indicePaginaEnCache(pagina int) int {
	time.Sleep(time.Millisecond * time.Duration(global.CpuConfig.CacheDelay)) 
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			return i
		}
	}
	return -1
}

func indicePaginaEnTLB(pagina int) int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == pagina {
			global.TLB[i].UltimoUso = lruCounter

			return i
		}
	}
	return -1
}

func indiceVacioTLB() int {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == -1 {
			return i
		}
	}
	return -1
}

func indiceVacioCACHE() int {
	time.Sleep(time.Millisecond * time.Duration(global.CpuConfig.CacheDelay)) 
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == -1 {
			return i
		}
	}
	return -1
}

func escribirCache(indice int, datos string, desplazamiento int){
	contenido := global.CACHE[indice].Contenido
	copy(contenido[desplazamiento:], []byte(datos))
	global.CACHE[indice].BitModificado = 1
	
}