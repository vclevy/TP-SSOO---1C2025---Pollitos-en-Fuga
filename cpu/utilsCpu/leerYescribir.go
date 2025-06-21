package utilsIo

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"time"
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
		if CacheHIT(nroPagina) {  //!! CACHE HIT
			indice := indicePaginaEnCache(nroPagina)
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Acción: LEER - Dirección Física: %d - Valor: %s", global.PCB_Actual.PID, direccionFisica, global.CACHE[indice].Contenido), log.INFO)
		} else {
			if tlbHabilitada {
				if TlbHIT(nroPagina) {
					marco = global.TLB[indice].Marco
					direccionFisica = MMU(desplazamiento, marco)
					contenidoLeido,_ := MemoriaLee(direccionFisica, tamanio)

					actualizarTLB(nroPagina, marco)
					actualizarCACHE(nroPagina, contenidoLeido)
				} else {
					marco = CalcularMarco()
					direccionFisica = marco * configMMU.Tamanio_pagina
					contenidoLeido,_ := MemoriaLee(direccionFisica, tamanio)
					actualizarTLB(nroPagina, marco)
					actualizarCACHE(nroPagina, contenidoLeido)
				}
			} else {
				marco = CalcularMarco()
				direccionFisica = marco * configMMU.Tamanio_pagina
				contenidoLeido,_ := MemoriaLee(direccionFisica, tamanio)
				actualizarCACHE(nroPagina, contenidoLeido)
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

func TlbHIT(pagina int) bool {
	for i := 0; i <= len(global.TLB)-1; i++ {
		if global.TLB[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB HIT - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo TLB HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - TLB MISS - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo TLB MISS
	return false
}

func CacheHIT(pagina int) bool {
	time.Sleep(global.CpuConfig.CacheDelay)
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == pagina {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Hit - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo CACHE HIT
			indice = i
			return true
		}
	}
	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Miss - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO) //todo CACHE MISS
	return false
}

func actualizarCACHE(pagina int, nuevoContenido string) {
	time.Sleep(global.CpuConfig.CacheDelay)
	var indicePisar int
	indice := indicePaginaEnCache(pagina)
	if indice == -1 { // no está la página en cache
		if indiceVacioCACHE() == -1 { // no hay espacio vacio en cachce
			indicePisar = AlgoritmoCACHE()
		} else {
			indicePisar = indiceVacioCACHE()
		}
		if global.CACHE[indicePisar].BitModificado == 1 {
			global.LoggerCpu.Log(fmt.Sprintf("PID: %d - Cache Add - Pagina: %d", global.PCB_Actual.PID, pagina), log.INFO)
			desalojar(indicePisar)
		}
		global.CACHE[indicePisar].NroPagina = pagina
		global.CACHE[indicePisar].Contenido = nuevoContenido
		global.CACHE[indicePisar].BitModificado = 0
	} else {
		global.CACHE[indice].Contenido = nuevoContenido
		global.CACHE[indice].BitModificado = 1
	}
}

func actualizarTLB(pagina int, marco int) {
	var indicePisar int
	indice := indicePaginaEnTLB(pagina)
	if indice == -1 { // no está la página
		if indiceVacioTLB() == -1 {
			indicePisar = AlgoritmoTLB()
			lruCounter++
		} else {
			indicePisar = indiceVacioTLB()
		}
		global.TLB[indicePisar].UltimoUso = lruCounter
		global.TLB[indicePisar].Marco = marco
		global.TLB[indicePisar].NroPagina = pagina

	} else {
		global.TLB[indice].Marco = marco
	}
}

func indicePaginaEnCache(pagina int) int {
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
	for i := 0; i <= len(global.CACHE)-1; i++ {
		if global.CACHE[i].NroPagina == -1 {
			return i
		}
	}
	return -1
}