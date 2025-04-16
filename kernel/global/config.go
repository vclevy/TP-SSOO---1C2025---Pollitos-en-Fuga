package global

import(
    log "github.com/sisoputnfrba/tp-golang/utils/logger"
    utils "github.com/sisoputnfrba/tp-golang/utils/config"
)

var KernelConfig *Config
var Logger *log.LoggerStruct

type Config struct {
    IPMemory          		string 		`json:"ip_memory"`
    Port_Memory         	int    		`json:"port_memory"`
    SchedulerAlgorithm 		string 		`json:"scheduler_algorithm"`
	ReadyIngressALgorithm 	string		`json:"ready_ingress_algorithm"`
	Alpha 					string		`json:"alpha"`
    SuspensionTime      	int    		`json:"suspension_time"`
    LogLevel          		string 		`json:"log_level"`
    Port_Kernel         	int    		`json:"port_kernel"`
	Log_file          		string 		`json:"log_file"`
	Ip_Kernel				string 		`json:"ip_kernel"`
}

func InitGlobal() {
	// 1. Cargar configuración desde archivo
	KernelConfig = utils.CargarConfig[Config]("config/config.json")

	// 2. Inicializar logger con lo que vino en la config
	Logger = log.ConfigurarLogger(KernelConfig.Log_file, KernelConfig.LogLevel)
    Logger.Log("Logger de Kernel inicializado", log.DEBUG)
}

