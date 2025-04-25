package estructuras

type PaqueteMemoria struct {
	PID int `json:"pid"`
	ArchivoPseudocodigo  	string  `json:"archivo_codigo"`
	TamanioProceso int `json:"tamanioProceso"`
}