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

type ConfiguracionMMU struct {
	Tamaño_página 			int 	`json:"tamaño_página"`
	Cant_entradas_tabla  	int     `json:"cant_entradas_tabla"`
	Cant_N_Niveles    		int     `json:"cant_N_Niveles"`
}

//*Syscalls

type Syscall_IO struct {
	IoSolicitada   string `json:"ioSolicitada"`
	TiempoEstimado int    `json:"tiempoEstimado"`
	PIDproceso     int    `json:"PIDproceso"`
}

type Syscall_Init_Proc struct {
	ArchivoInstrucciones	string `json:"archivoInstrucciones"`
	Tamanio					int    `json:"tamanio"`
	PIDproceso				int    `json:"PIDproceso"`
}

/* type Syscall_Dump_Memory struct {	
	PIDproceso     int    `json:"PIDproceso"`
}

type Syscall_Exit struct {
	PIDproceso     int    `json:"PIDproceso"`
} */


// para syscall init proc creo que es lo mismo que Paquete memoria mas arriba
// exit no recibe parametros solo necesito saber el proceso que la invoco asi q solo pasame el pid creo
// dump memory solo neceisto saber el proceso que la invoco tmb (o sea su pid)
/*CONSIGNA PARA LAS SYSCALLS (kernel)
	Dentro de las syscalls que se pueden atender referidas a procesos, tendremos las instrucciones INIT_PROC y EXIT.
INIT_PROC, esta syscall recibirá 2 parámetros de la CPU, el primero será el nombre del archivo de pseudocódigo que deberá ejecutar el proceso y el segundo parámetro es el tamaño del proceso en Memoria. El Kernel creará un nuevo PCB y lo dejará en estado NEW, esta syscall no implica cambio de estado, por lo que el proceso que llamó a esta syscall, inmediatamente volverá a ejecutar en la CPU.
EXIT, esta syscall no recibirá parámetros y se encargará de finalizar el proceso que la invocó, siguiendo lo descrito anteriormente para Finalización de procesos.

En este apartado solamente se tendrá la instrucción DUMP_MEMORY. Esta syscall le solicita a la memoria, junto al PID que lo solicitó, que haga un Dump del proceso.
Esta syscall bloqueará al proceso que la invocó hasta que el módulo memoria confirme la finalización de la operación, en caso de error, el proceso se enviará a EXIT. Caso contrario, se desbloquea normalmente pasando a READY.

*/
type TareaDeIo struct {
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}

type FinDeIO struct {
	Tipo    string `json:"tipo"` 
}