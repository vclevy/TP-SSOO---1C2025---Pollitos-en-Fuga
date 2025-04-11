package global

import(
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var KernelConfig *Config
var Logger *logger.LoggerStruct


type Config struct {
    IPMemory           string 	`json:"ip_memory"`
    PortMemory         int    	`json:"port_memory"`
    PortKernel         int    	`json:"port_kernel"`
    SchedulerAlgorithm string 	`json:"scheduler_algorithm"`
    SuspensionTime     int    	`json:"suspension_time"`
    LogLevel           string 	`json:"log_level"`
	Log_file           string 	`json:"log_file"`
}
