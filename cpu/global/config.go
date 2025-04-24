package global

import (
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"fmt"
)

var CpuConfig *Config
var LoggerCpu *logger.LoggerStruct

type Config struct {
    Ip_Memory           	string 	`json:"ip_memory"`
	Ip_Cpu           	string 	`json:"ip_cpu"`
	Ip_Kernel 			string	`json:"ip_kernel"`
    Port_Memory         int    	`json:"port_memory"`
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

func InitGlobal(idCPU string) {	
	CpuID = idCPU
	// 1. Cargar configuración desde archivo
	CpuConfig = utils.CargarConfig[Config]("config/config.json")

	// 2. crear el archivo Log correspondiente a la CPU 
	logFileName := fmt.Sprintf("logs/%s.log", idCPU)

	// 4. Inicializar archivo logger con ese nombre 
	LoggerCpu = logger.ConfigurarLogger(logFileName, CpuConfig.LogLevel)

	// 5. avisar que fue inicializado
    LoggerCpu.Log("Logger de CPU inicializado", logger.DEBUG)
}