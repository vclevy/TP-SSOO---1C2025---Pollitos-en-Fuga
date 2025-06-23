package global

import (
	"fmt"
	"time"
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"os"
	log "github.com/sisoputnfrba/tp-golang/utils/logger")

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

var CpuID string
var Interrupcion bool
var PCB_Actual estructuras.PCB
var Motivo string
var Rafaga float64

var CacheHabilitada bool
var TlbHabilitada bool
var PCB_Interrupcion estructuras.PCB
var TamPagina int

var TLB []estructuras.DatoTLB
var CACHE []estructuras.DatoCACHE

func InitGlobal(idCPU string) {
	CpuID = idCPU
	// 1. Cargar configuración desde archivo
	CpuConfig = utils.CargarConfig[Config]("config/config.json")

	os.MkdirAll("logs", os.ModePerm)

	// 2. Crear el archivo Log correspondiente a la CPU
	logFileName := fmt.Sprintf("logs/%s.log", idCPU)

	// 4. Inicializar archivo logger con ese nombre
	LoggerCpu = log.ConfigurarLogger(logFileName, CpuConfig.LogLevel)

	// 5. Avisar que fue inicializado
	LoggerCpu.Log("Logger de CPU inicializado", log.DEBUG)

	CacheHabilitada = CpuConfig.CacheEntries > 0
	TlbHabilitada =  CpuConfig.TlbEntries > 0

	

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
			UltimoUso: -1,
		}
	}
}

func InicializarCACHE() {
	CACHE = make([]estructuras.DatoCACHE, CpuConfig.CacheEntries)
	for i := range CACHE {
		CACHE[i] = estructuras.DatoCACHE{
			BitModificado: -1,
			NroPagina:     -1,
			Contenido:		make([]byte, 64),//ver
			BitUso:        -1,
		}
	}
}

