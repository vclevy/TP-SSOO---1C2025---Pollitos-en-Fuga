package global

import (
	"fmt"

	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var CpuConfig *Config
var LoggerCpu *logger.LoggerStruct

type Config struct {
    Ip_Memoria        	string `json:"ip_memory"`
	Ip_Cpu           	string 	`json:"ip_cpu"`
	Ip_Kernel 			string	`json:"ip_kernel"`
    Port_Memoria       	int    `json:"port_memory"`
    Port_Cpu 		    int 	`json:"port_cpu"`
    Port_Kernel         int    	`json:"port_kernel"`
    TlbEntries     		int    	`json:"tlb_entries"`
    TlbReplacement      string 	`json:"tlb_replacement"`
	CacheEntries        int 	`json:"cache_entries"`
	CacheReplacement	int		`json:"cache_replacement"`
	CacheDelay			int		`json:"cache_delay"`
	LogLevel			string	`json:"log_level"`
	LogFile				string  `json:"log_file"`
}

var CpuID string
var Interrupcion bool
var PCB_Actual estructuras.PCB
var Motivo string
var Rafaga float64

var TLB []estructuras.DatoTLB
var CACHE []estructuras.DatoCACHE

func InitGlobal(idCPU string) {	
	CpuID = idCPU
	// 1. Cargar configuración desde archivo
	CpuConfig = utils.CargarConfig[Config]("config/config.json")

	// 2. Crear el archivo Log correspondiente a la CPU 
	logFileName := fmt.Sprintf("logs/%s.log", idCPU)

	// 4. Inicializar archivo logger con ese nombre 
	LoggerCpu = logger.ConfigurarLogger(logFileName, CpuConfig.LogLevel)

	// 5. Avisar que fue inicializado
    LoggerCpu.Log("Logger de CPU inicializado", logger.DEBUG)
}