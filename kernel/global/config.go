package global

// import (
// 	"fmt"
// 	"os"
// 	"sync")

// func InitGlobal() {
// 	args := os.Args[1:]
// 	if len(args) <= 2 {
// 		fmt.Println("Argumentos esperados para iniciar el servidor: ENV=dev | prod CONFIG=config_path")
// 		os.Exit(1)
// 	}
// 	env := args[0]
// 	archivoConfiguracion := args[1]

// 	Logger = log.ConfigurarLogger(KernelLog, env)
// 	KernelConfig = config.CargarConfig[Config](archivoConfiguracion)
// 	EstadoListo = make(map[int][]estructuras.TCB)
// 	EstadoListo[0] = []estructuras.TCB{}
// 	PrioridadesEnSistema = append(PrioridadesEnSistema, 0)
// 	// var ProcesoEjecutando = 0 // no seria variable global (?)

// }