

package global

import(
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)


var IoConfig *Config
var Logger *logger.LoggerStruct

type Config struct {
	IPKernel           	string 	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    Port_Io 		   	int 	`json:"port_io"`
	LogLevel			string	`json:"log_level"`
	Log_File			string  `json:"log_file"`	
	}	
