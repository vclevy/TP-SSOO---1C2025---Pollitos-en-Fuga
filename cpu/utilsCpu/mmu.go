package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"math"
	"net/http"
)

func ConfigMMU() error {
	url := fmt.Sprintf("http://%s:%d/configuracionMMU", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria)
	resp, err := http.Get(url)

	if err != nil {
		global.LoggerCpu.Log("Error al conectar con Memoria:", log.ERROR)
		return err
	}
	defer resp.Body.Close() //cierra automáticamente el cuerpo de la respuesta HTTP

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerCpu.Log("Error leyendo respuesta de Memoria:", log.ERROR)
		return err
	}

	err = json.Unmarshal(body, &configMMU) // convierto el JSON que recibi de Memoria y lo guardo en el struct configMMU.
	if err != nil {
		global.LoggerCpu.Log("Error parseando JSON de configuración:", log.ERROR)
		return err
	}

	return nil
}

func armarListaEntradas(nroPagina int) []int {
	cantNiveles := configMMU.Cant_N_Niveles
	cantEntradas := configMMU.Cant_entradas_tabla

	entradas := make([]int, cantNiveles)

	for i := 1; i <= cantNiveles; i++ {
		entradas[i-1] = int(math.Floor(float64(nroPagina)/math.Pow(float64(cantEntradas), float64(cantNiveles-i)))) % cantEntradas
	}
	return entradas
}

func CalcularMarco() int {
	listaEntradas := armarListaEntradas(nroPagina)

	accederTabla := estructuras.AccesoTP{
		PID:      global.PCB_Actual.PID,
		Entradas: listaEntradas,
	}

	marco := pedirMarco(accederTabla)

	global.LoggerCpu.Log(fmt.Sprintf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", global.PCB_Actual.PID, nroPagina, marco), log.INFO)

	return marco
}

func MMU(desplazamiento int, marco int) int {
	direccionFisica = marco*configMMU.Tamanio_pagina + desplazamiento
	return direccionFisica
}

func pedirMarco(estructuras.AccesoTP) int {
	var accesoTP estructuras.AccesoTP

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

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria de forma exitosa", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var marco int
	err = json.Unmarshal(body, &marco)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return -1
	}
	return marco
}
