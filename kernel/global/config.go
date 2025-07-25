package global

import (
	"sync"
	"time"

	utils "github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
)

var ConfigKernel *Config
var LoggerKernel *logger.LoggerStruct
var UltimoPID int = 0
var MutexUltimoPID sync.Mutex

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
    MemoriaRequerida  int
    ArchivoPseudo     string
    EstimacionRafaga  float64
	InstanteInicio time.Time
	UltimaRafagaReal float64
	UltimaRafagaEstimada float64
}

func NuevoPCB() *PCB {
	pid := UltimoPID

    return &PCB{
        PID:          pid,
        PC:           0,
        ME:           make(map[string]int),
        MT:           make(map[string]int),
        UltimoEstado: "",
        InicioEstado: time.Now(),
    }
}
type Config struct {
	IPMemory              string  `json:"ip_memory"`
	Port_Memory           int     `json:"port_memory"`
	SchedulerAlgorithm    string  `json:"scheduler_algorithm"`
	ReadyIngressAlgorithm string  `json:"ready_ingress_algorithm"`
	Alpha                 float64 `json:"alpha"`
	SuspensionTime        int     `json:"suspension_time"`
	LogLevel              string  `json:"log_level"`
	Port_Kernel           int     `json:"port_kernel"`
	Log_file              string  `json:"log_file"`
	Ip_Kernel             string  `json:"ip_kernel"`
	InitialEstimate       int     `json:"initial_estimate"`
}

func InitGlobal(archivoConfig string) {
	ConfigKernel = utils.CargarConfig[Config](archivoConfig)

	LoggerKernel = logger.ConfigurarLogger(ConfigKernel.Log_file, ConfigKernel.LogLevel)
	LoggerKernel.Log("Logger de Kernel inicializado", logger.DEBUG)

	InicioPlanificacionLargoPlazo = make(chan struct{})
}

var InicioPlanificacionLargoPlazo chan struct{}

var ColaNew []*Proceso
var ColaReady []*Proceso
var ColaSuspReady []*Proceso
var ColaExecuting []*Proceso
var ColaBlocked []*Proceso
var ColaSuspBlocked []*Proceso
var ColaExit []*Proceso

var MutexNew sync.Mutex
var MutexReady sync.Mutex
var MutexSuspReady sync.Mutex
var MutexExecuting sync.Mutex
var MutexBlocked sync.Mutex
var MutexSuspBlocked sync.Mutex
var MutexExit sync.Mutex

var (
    NotifySuspReady = make(chan struct{}, 1)
    NotifyNew       = make(chan struct{}, 1)
)

func AgregarASuspReady(p *Proceso) {
    MutexSuspReady.Lock()
    ColaSuspReady = append(ColaSuspReady, p)
    MutexSuspReady.Unlock()
}

func AgregarANew(p *Proceso) {
	MutexNew.Lock()
	ColaNew = append(ColaNew, p)
	MutexNew.Unlock()

	select {
	case NotifyNew <- struct{}{}:
	default: 
	}
}

var NotifyReady = make(chan struct{}, 100)

func AgregarAExecuting(p *Proceso) {
	MutexExecuting.Lock()
	ColaExecuting = append(ColaExecuting, p)
	MutexExecuting.Unlock()
}
func AgregarABlocked(p *Proceso) {
	MutexBlocked.Lock()
	ColaBlocked = append(ColaBlocked, p)
	MutexBlocked.Unlock()
}

func AgregarASuspBlocked(p *Proceso) {
	MutexSuspBlocked.Lock()
	ColaSuspBlocked = append(ColaSuspBlocked, p)
	MutexSuspBlocked.Unlock()
}

func AgregarAExit(p *Proceso) {
	MutexExit.Lock()
	ColaExit = append(ColaExit, p)
	MutexExit.Unlock()
}

// CPU
type CPU struct {
	ID                string
	Puerto            int
	IP                string
	ProcesoEjecutando *PCB
	CanalProceso      chan *Proceso
}

var CPUsConectadas []*CPU
var MutexCPUs sync.Mutex

//IO

var IOListMutex sync.RWMutex
type IOData = estructuras.IOData
var IOConectados []*IODevice
type ProcesoIO struct {
	Proceso   *Proceso
	TiempoUso int
}
type IODevice struct {
	Nombre       string
	IP           string
	Puerto       int
	Ocupado      bool
	ProcesoEnUso *ProcesoIO
	ColaEspera   []*ProcesoIO 
	Mutex        sync.Mutex
}

func EliminarProcesoDeCola(cola *[]*Proceso, pid int) bool {
	for i, p := range *cola {
		if p.PID == pid {
			*cola = append((*cola)[:i], (*cola)[i+1:]...)
			return true
		}
	}
	return false
}

func NotificarReady() {
	select {
	case NotifyReady <- struct{}{}:
	default:
	}
}