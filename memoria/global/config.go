package global

import(
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"sync"
)
var ConfigMemoria *Config
var LoggerMemoria *logger.LoggerStruct

type Config struct {
	Port_Memory      int      `json:"port_memory"`
	IPMemory		string 		`json:"ip_memory"`
	Memory_size      int      `json:"memory_size"`
	Page_Size		 int  	  `json:"page_size"`
	Entries_per_page int      `json:"entries_per_page"`
	Number_of_levels int      `json:"number_of_levels"`
	Memory_delay     int      `json:"memory_delay"`
	Swapfile_path    string   `json:"swapfile_path"`
	Swap_delay 		 int      `json:"swap_delay"`
	Log_Level        string   `json:"log_level"`
	Dump_path		 string   `json:"dump_path"`
	Log_file         string   `json:"log_file"`
	Scripts_Path string `json:"scripts_path"`
	//"scripts_file": "/home/utnso/scripts/" LO CAMBIE PARA PRUEBAS
}

func InitGlobal(configPath string) {
	ConfigMemoria = utils.CargarConfig[Config](configPath)
	LoggerMemoria = logger.ConfigurarLogger(ConfigMemoria.Log_file, ConfigMemoria.Log_Level)
	LoggerMemoria.Log("Logger de Memoria inicializado", logger.DEBUG)
}


var MutexMemoriaUsuario sync.Mutex
var MutexSwap sync.Mutex
var MutexInstrucciones sync.RWMutex
var MutexMarcos sync.Mutex
var MutexMetricas sync.Mutex
