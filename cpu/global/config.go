package global

import (
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var CpuConfig *Config
var LoggerCPU *logger.LoggerStruct

type Config struct {
    IPMemory           	int 	`json:"ip_memory"`
	IPCpu           	int 	`json:"ip_cpu"`
    Port_Memory         int    	`json:"port_memory"`
	IPKernel 			string	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    Port_CPU 		    int 	`json:"port_cpu"`
    TlbEntries     		int    	`json:"tlb_entries"`
    TlbReplacement      string 	`json:"tlb_replacement"`
	CacheEntries        int 	`json:"cache_entries"`
	CacheReplacement	int		`json:"cache_replacement"`
	CacheDelay			int		`json:"cache_delay"`
	LogLevel			string	`json:"log_level"`
	LogFile				string  `json:"log_file"`	
}	

func InitGlobal() {
	// 1. Cargar configuración desde archivo
	CpuConfig = utils.CargarConfig[Config]("config/config.json")

	// 2. Inicializar logger con lo que vino en la config
	LoggerCPU = logger.ConfigurarLogger(CpuConfig.LogFile, CpuConfig.LogLevel)
    LoggerCPU.Log("Logger de CPU inicializado", logger.DEBUG)
}