package global

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var CpuConfig *Config
var LoggerCpu *log.LoggerStruct

type Config struct {
	Ip_Memoria       string        `json:"ip_memory"`
	Ip_Cpu           string        `json:"ip_cpu"`
	Ip_Kernel        string        `json:"ip_kernel"`
	Port_Memoria     int           `json:"port_memory"`
	Port_Cpu         int           `json:"port_cpu"`
	Port_Kernel      int           `json:"port_kernel"`
	TlbEntries       int           `json:"tlb_entries"`
	TlbReplacement   string        `json:"tlb_replacement"`
	CacheEntries     int           `json:"cache_entries"`
	CacheReplacement string        `json:"cache_replacement"`
	CacheDelay       time.Duration `json:"cache_delay"`
	LogLevel         string        `json:"log_level"`
	LogFile          string        `json:"log_file"`
}

var IO_Request estructuras.Syscall_IO
/* var Init_Proc estructuras.Syscall_Init_Proc */
var CpuID string
var Interrupcion bool
var PCB_Actual *estructuras.PCB
var Motivo string
var Rafaga float64

var CacheHabilitada bool
var TlbHabilitada bool
var PCB_Interrupcion *estructuras.PCB
var TamPagina int
var ConfigMMU estructuras.ConfiguracionMMU

var TLB []estructuras.DatoTLB
var CACHE []estructuras.DatoCACHE

func InitGlobal(idCPU string, configPath string) {
	CpuID = idCPU

	// 1. Cargar configuración desde archivo recibido por parámetro
	CpuConfig = utils.CargarConfig[Config](configPath)

	os.MkdirAll("logs", os.ModePerm)

	// 2. Crear el archivo Log correspondiente a la CPU
	logFileName := fmt.Sprintf("logs/%s.log", idCPU)

	// 3. Inicializar archivo logger
	LoggerCpu = log.ConfigurarLogger(logFileName, CpuConfig.LogLevel)
	LoggerCpu.Log("Logger de CPU inicializado", log.INFO)

	CacheHabilitada = CpuConfig.CacheEntries > 0
	TlbHabilitada = CpuConfig.TlbEntries > 0

	if err := CargarConfigMMU(); err != nil {
		LoggerCpu.Log("Error en ConfigMMU: "+err.Error(), log.ERROR)
	}

	if CacheHabilitada {
		InicializarCACHE()
	}
	if TlbHabilitada {
		InicializarTLB()
	}
}

func InicializarTLB() {
	TLB = make([]estructuras.DatoTLB, CpuConfig.TlbEntries)
	for i := range TLB {
		TLB[i] = estructuras.DatoTLB{
			NroPagina: -1,
			Marco:     -1,
			UltimoUso: 0,
		}
	}
}

func InicializarCACHE() {
	CACHE = make([]estructuras.DatoCACHE, CpuConfig.CacheEntries)
	for i := range CACHE {
		CACHE[i] = estructuras.DatoCACHE{
			BitModificado: -1,
			NroPagina:     -1,
			Contenido:     make([]byte, ConfigMMU.Tamanio_pagina),
			BitUso:        -1,
		}
	}
}

func CargarConfigMMU() error {
	url := fmt.Sprintf("http://%s:%d/configuracionMMU", CpuConfig.Ip_Memoria, CpuConfig.Port_Memoria)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		LoggerCpu.Log("Error al conectar con Memoria:", log.ERROR)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LoggerCpu.Log("Error leyendo respuesta de Memoria:", log.ERROR)
		return err
	}

	LoggerCpu.Log("JSON recibido de Memoria: "+string(body), log.DEBUG)

	err = json.Unmarshal(body, &ConfigMMU)
	if err != nil {
		LoggerCpu.Log("Error parseando JSON de configuracion: "+err.Error(), log.ERROR)
		return err
	}
	LoggerCpu.Log(fmt.Sprintf("Entradas tabla %d", ConfigMMU.Cant_entradas_tabla), log.DEBUG)
	LoggerCpu.Log(fmt.Sprintf("tamanio pagina %d", ConfigMMU.Tamanio_pagina), log.DEBUG)
	LoggerCpu.Log(fmt.Sprintf("cantidad niveles %d", ConfigMMU.Cant_N_Niveles), log.DEBUG)

	TamPagina = ConfigMMU.Tamanio_pagina
	return nil
}