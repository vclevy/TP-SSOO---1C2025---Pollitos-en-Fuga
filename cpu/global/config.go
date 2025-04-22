package global

import (
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"fmt"
)

var CpuConfig *Config
var LoggerCpu *logger.LoggerStruct

type Config struct {
    IPMemory           	string 	`json:"ip_memory"`
	IPCpu           	string 	`json:"ip_cpu"`
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
	/* Identificador		string	`json:"-"` */
}

var CpuID string

func InitGlobal(idCPU string) {	
	CpuID = idCPU
	// 1. Cargar configuración desde archivo
	CpuConfig = utils.CargarConfig[Config]("config/config.json")

	logFileName := fmt.Sprintf("logs/%s.log", idCPU)

	// 4. Inicializar logger con ese archivo
	LoggerCpu = logger.ConfigurarLogger(logFileName, CpuConfig.LogLevel)
    LoggerCpu.Log("Logger de CPU inicializado", logger.DEBUG)
}