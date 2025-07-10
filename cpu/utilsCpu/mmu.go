package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

func armarListaEntradas(nroPagina int) []int {
	cantNiveles := global.ConfigMMU.Cant_N_Niveles
	cantEntradas := global.ConfigMMU.Cant_entradas_tabla

	entradas := make([]int, cantNiveles)

	for i := 1; i <= cantNiveles; i++ {
		/* global.LoggerCpu.Log(fmt.Sprintf("Dividiendo para nivel %d", i), log.DEBUG) */ 
	
		exponente := cantNiveles - i
		if exponente < 0 {
			global.LoggerCpu.Log(fmt.Sprintf("ERROR: Exponente negativo. Nivel: %d, cantNiveles: %d", i, cantNiveles), log.ERROR)
			return nil // o panic, o manejo de error
		}
	
		divisor := math.Pow(float64(cantEntradas), float64(exponente))
		if divisor == 0 {
			global.LoggerCpu.Log("ERROR: División por cero en armarListaEntradas", log.ERROR)
			return nil
		}
	
		entradas[i-1] = int(math.Floor(float64(nroPagina)/divisor)) % cantEntradas
	}
	
	return entradas
}

func CalcularMarco(numeroPagina int) int {
	if global.TlbHabilitada { //TLB HABILITADA
		if TlbHIT(numeroPagina) { // CASO: esta en la TLB
			indiceHIT := indicePaginaEnTLB(numeroPagina)
			global.TLB[indiceHIT].UltimoUso = lruCounter
			return global.TLB[indiceHIT].Marco
		} else { // CASO: NO esta en la TLB
			return actualizarTLB()
		}
	}
	//TLB NO ESTA HABILITADA
	return BuscarMarcoEnMemoria(numeroPagina)
}

func BuscarMarcoEnMemoria(nroPagina int) int {
	listaEntradas := armarListaEntradas(nroPagina)

	accederTabla := estructuras.AccesoTP{
		PID:      global.PCB_Actual.PID,
		Entradas: listaEntradas,
	}

	marco := pedirMarco(accederTabla)

	return marco
}

func MMU(desplazamiento int, marco int) int {
	direccionFisica = marco* global.ConfigMMU.Tamanio_pagina + desplazamiento
	return direccionFisica
}

func pedirMarco(accesoTP estructuras.AccesoTP) int {

	jsonData, err := json.Marshal(accesoTP)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return -1
	}

	url := fmt.Sprintf("http://%s:%d/pedirMarco", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                                //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: "+err.Error(), log.ERROR)
		return -1
	}
	defer resp.Body.Close() //se cierra la conexión

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var marco int
	err = json.Unmarshal(body, &marco)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return -1
	}

	global.LoggerCpu.Log(fmt.Sprintf("\033[36mPID: %d - OBTENER MARCO - Página: %d - Marco: %d\033[0m", global.PCB_Actual.PID, nroPagina, marco), log.INFO) //!! Obtener Marco - logObligatorio

	return marco
}
