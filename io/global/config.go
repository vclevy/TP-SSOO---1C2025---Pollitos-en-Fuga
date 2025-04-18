

package global

import(
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
)


var IoConfig *Config
var LoggerIo *logger.LoggerStruct

type Config struct {
	IPKernel           	string 	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    Port_Io 		   	int 	`json:"port_io"`
	IPIo				string 	`json:"ip_io"`
	LogLevel			string	`json:"log_level"`
	Log_File			string  `json:"log_file"`	
	}	

func InitGlobal () {
	IoConfig = utils.CargarConfig[Config]("config/config.json")
	LoggerIo = logger.ConfigurarLogger(IoConfig.Log_File, IoConfig.LogLevel)
	LoggerIo.Log("Logger de io inicializado", logger.DEBUG)
}