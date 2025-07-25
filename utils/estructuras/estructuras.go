package estructuras
type PaqueteMemoria struct {
	PID                 int    `json:"pid"`
	ArchivoPseudocodigo string `json:"archivoPseudocodigo"`
	TamanioProceso      int    `json:"tamanioProceso"`
}

type IOData struct {
	IP     string
	Puerto int
	Cola   []int
	PID    int
}
type MensajeIO struct {
	NombreIO string `json:"nombre_io"`
	Evento   string `json:"evento"` 
	PID      int    `json:"pid"`    
	IP       string `json:"ip"`    
	Puerto   int    `json:"puerto"` 
}

type PaqueteHandshakeIO struct {
	NombreIO string `json:"nombreIO"`
	IPIO     string `json:"ipio"`
	PuertoIO int    `json:"puertoIO"`
}

type ConfiguracionMMU struct {
	Tamanio_pagina      int `json:"tamanio_pagina"`
	Cant_entradas_tabla int `json:"cant_entradas_tabla"`
	Cant_N_Niveles      int `json:"cant_N_Niveles"`
}
type Syscall_IO struct {
	IoSolicitada   string `json:"ioSolicitada"`
	TiempoEstimado int    `json:"tiempoEstimado"`
	PIDproceso     int    `json:"PIDproceso"`
}

type Syscall_Init_Proc struct {
	PID                 int    `json:"pid"`
	ArchivoInstrucciones string `json:"archivoInstrucciones"`
	Tamanio              int    `json:"tamanio"`
}
type TareaDeIo struct {
	PID            int `json:"pid"`
	TiempoEstimado int `json:"tiempo_estimado"`
}
type FinDeIO struct {
	Tipo string `json:"tipo"`
	PID  int    `json:"PID"`
}
type HandshakeConCPU struct {
	ID     string
	Puerto int
	IP     string
}
type SolicitudDump struct {
	PID int `json:"pid"`
}

type RespuestaCPU struct {
	PID        int       `json:"PID"`
	PC         int       `json:"PC"`
	Motivo     string    `json:"Motivo"`
	RafagaReal float64   `json:"RafagaReal"`
	IO         Syscall_IO `json:"IO"` 
/* 	InitProc 	Syscall_Init_Proc `json:"InicProc"`// <-- nuevo campo */
}

type PCB struct {
	PID int `json:"pid"`
	PC  int `json:"pc"`
}

type AccesoTP struct {
	PID      int   `json:"pid"`
	Entradas []int `json:"entradas"`
}

type PedidoREAD struct {
	PID             int `json:"pid"`
	DireccionFisica int `json:"direccion_fisica"`
	Tamanio         int `json:"tamanio"`
}

type PedidoWRITE struct {
	PID             int    `json:"pid"`
	DireccionFisica int    `json:"direccion_fisica"`
	Datos           []byte `json:"datos"`
}

type DatoTLB struct {
	NroPagina int
	Marco     int
	UltimoUso int
}

type DatoCACHE struct {
	BitModificado int
	NroPagina     int
	Contenido     []byte
	BitUso        int // 0 o 1
}