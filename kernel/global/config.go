package global

import (
	"fmt"
	"os"

	config "github.com/sisoputnfrba/tp-golang/utils/config"
	// estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
)

const KernelLog = "./kernel.log"
type Config struct {
    IPMemory           string `json:"ip_memory"`
    PortMemory         int    `json:"port_memory"`
    PortKernel         int    `json:"port_kernel"`
    SchedulerAlgorithm string `json:"scheduler_algorithm"`
    SuspensionTime     int    `json:"suspension_time"`
    LogLevel           string `json:"log_level"`
}

var KernelConfig *Config
var Logger *log.LoggerStruct

func IniciarKernel(){

 args := os.Args[1:]

	if len(args) <= 2 {
		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
		os.Exit(1)
	}
	env := args[0]
	archivoConfiguracion := args[1]

	Logger = log.ConfigurarLogger(KernelLog, env)
	KernelConfig = config.CargarConfig[Config](archivoConfiguracion)
}
