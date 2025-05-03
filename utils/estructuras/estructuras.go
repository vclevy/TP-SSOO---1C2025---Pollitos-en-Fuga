package estructuras

type PaqueteMemoria struct {
	PID int `json:"pid"`
	ArchivoPseudocodigo  	string  `json:"archivo_codigo"`
	TamanioProceso int `json:"tamanioProceso"`
}

type IOData struct {
	IP     string
	Puerto int
	Cola   []int
	PID int 
}
type MensajeIO struct {
	NombreIO string `json:"nombre_io"`
	Evento   string `json:"evento"`   // "registro", "fin", "desconexion"
	PID      int    `json:"pid"`      // Opcional, solo si es fin
	IP       string `json:"ip"`       // Solo para registro
	Puerto   int    `json:"puerto"`   // Solo para registro
}

type PaqueteHandshakeIO struct {
	NombreIO 		string	 `json:"nombreIO"`
	IPIO 			string    	 `json:"ipio"`
	PuertoIO    	int      `json:"puertoIO"`
}

type Instruccion struct {
	Pid		int		`json:"Pid"`
	Pc		int		`json:"Pc"`
}

type Syscall struct {
	InstruccionSyscall	Instruccion `json:"Instruccion"`
	IdCpu 	string 	`json:"IdCPU"`
}