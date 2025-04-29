package global

import (
	"time"

	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var ConfigKernel *Config
var LoggerKernel *logger.LoggerStruct
var UltimoPID int = 0

type PCB struct {
	PID          int
	PC           int
	ME           map[string]int // Métricas de Estado (ej: "Ready": 3, "Running": 5)
	MT           map[string]int // Métricas de Tiempo por Estado (ej: "Ready": 120, "Running": 300)
	UltimoEstado string
	InicioEstado time.Time
}
type Proceso struct {
	PCB
	MemoriaRequerida int
	ArchivoPseudo    string
	EstimacionRafaga float64
}


func NuevoPCB() *PCB {
	pid := UltimoPID
	UltimoPID++ 

	return &PCB{
		PID: pid,
		PC:  0,
		ME:  make(map[string]int),
		MT:  make(map[string]int),
	}
}



type Config struct {
    IPMemory          		string 		`json:"ip_memory"`
    Port_Memory         	int    		`json:"port_memory"`
    SchedulerAlgorithm 		string 		`json:"scheduler_algorithm"`
	ReadyIngressALgorithm 	string		`json:"ready_ingress_algorithm"`
	Alpha 					float64		`json:"alpha"`
    SuspensionTime      	int    		`json:"suspension_time"`
    LogLevel          		string 		`json:"log_level"`
    Port_Kernel         	int    		`json:"port_kernel"`
	Log_file          		string 		`json:"log_file"`
	Ip_Kernel				string 		`json:"ip_kernel"`
	EstimacionRafaga		float64		`json:"estimacion_rafaga"`
}

func InitGlobal() {
	// 1. Cargar configuración desde archivo
	ConfigKernel = utils.CargarConfig[Config]("config/config.json")

	// 2. Inicializar logger con lo que vino en la config
	LoggerKernel = logger.ConfigurarLogger(ConfigKernel.Log_file, ConfigKernel.LogLevel)
    LoggerKernel.Log("Logger de Kernel inicializado", logger.DEBUG)

	// 3. Inicializar canal de sincronización para planificación
	InicioPlanificacionLargoPlazo = make(chan struct{})
}

var InicioPlanificacionLargoPlazo chan struct{}

var ColaNew []Proceso
var ColaReady []Proceso
var ColaSuspReady []Proceso
var ColaExecuting []Proceso
var ColaBlocked []Proceso
var ColaSuspBlocked []Proceso
var ColaExit []Proceso

//CPU
var CantidadCPUsTotales int 
var CantidadCPUsOcupadas int // *OBVIO Q ESTO no va a terminar así, es para q compile hayCpuDisponible


//IO

type IOData = estructuras.IOData
var IOConectados []IODevice

type IODevice struct {
    Nombre      string           // ej. "impresora"
    IP          string           // IP del módulo IO
    Puerto      int              // Puerto del módulo IO
    Ocupado     bool             // ¿Está actualmente ocupado?
    ProcesoEnUso *Proceso        // Proceso que está usando la IO (nil si está libre)
    ColaEspera  []Proceso        // Lista de procesos esperando esta IO
}



