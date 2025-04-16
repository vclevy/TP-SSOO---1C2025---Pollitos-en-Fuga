package global

import(
    // "os"
    // "fmt"
    log "github.com/sisoputnfrba/tp-golang/utils/logger"
    utils "github.com/sisoputnfrba/tp-golang/utils/config"
)

var KernelConfig *Config
var Logger *log.LoggerStruct

type Config struct {
    IPMemory           string 	`json:"ip_memory"`
    Port_Memory         int    	`json:"port_memory"`
    Port_Kernel         int    	`json:"port_kernel"`
    SchedulerAlgorithm string 	`json:"scheduler_algorithm"`
    SuspensionTime     int    	`json:"suspension_time"`
	ReadyIngressALgorithm string		`json:"ready_ingress_algorithm"`
	Alpha 				string		`json:"alpha"`
    LogLevel           string 	`json:"log_level"`
	Log_file           string 	`json:"log_file"`
}

func InitGlobal() {
	// 1. Cargar configuración desde archivo
	KernelConfig = utils.CargarConfig[Config]("config/config.json")

	// 2. Inicializar logger con lo que vino en la config
	Logger = log.ConfigurarLogger(KernelConfig.Log_file, KernelConfig.LogLevel)
    Logger.Log("Logger de Kernel inicializado", log.DEBUG)
}




//go run ../kernel.go dev config/kernel.json

/* 
global.KernelConfig =  utils.CargarConfig[global.Config]("config/config.json")
	
	puertoMemoria := strconv.Itoa(global.KernelConfig.Port_Memory) 
	url := "http://localhost:"+ puertoMemoria+"/escribir" 
	body := []byte("hola desde kernel")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
	fmt.Println("Error al mandar mensaje a Memoria:", err)
	return
	 }
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)
	// 2. Inicializar logger
	global.Logger = logger.ConfigurarLogger(global.KernelConfig.Log_file, global.KernelConfig.LogLevel)
	defer global.Logger.CloseLogger()
	global.Logger.Log("Logger de memoria inicializado", logger.DEBUG) */