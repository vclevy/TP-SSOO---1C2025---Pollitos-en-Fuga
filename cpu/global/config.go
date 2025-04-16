package global

import (
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var CpuConfig *Config
var LoggerCpu *logger.LoggerStruct


type Config struct {
    IPMemory           	int 	`json:"ip_memory"`
	IPCpu           	int 	`json:"ip_cpu"`
    Port_Memory         int    	`json:"port_memory"`
	IPKernel 			string	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    PortCPU 		    int 	`json:"port_cpu"`
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
	LoggerCpu = logger.ConfigurarLogger(CpuConfig.LogFile, CpuConfig.LogLevel)
    LoggerCpu.Log("Logger de CPU inicializado", logger.DEBUG)
}



// "port_cpu": 8004,
// "ip_memory": "127.0.0.1",
// "port_memory": 8002,
// "ip_kernel": "127.0.0.1",
// "port_kernel": 8002,
// "tlb_entries": 15,
// "tlb_replacement": "TLB",
// "cache_entries": 10,
// "cache_replacement": "CLOCK",
// "cache_delay": 10,
// "log_level": "DEBUG"