package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	//"github.com/sisoputnfrba/tp-golang/utils/paquetes"
	"fmt"
)

//CONEXIÓN;
//[KERNEL] ➜ Cliente (conectado a) [MEMORIA]
//[CPU]    ➜ Cliente (conectado a) [MEMORIA]

func main() {

	global.InitGlobal()
	defer global.LoggerMemoria.CloseLogger()	
	
	// configurar logger e inicializar config
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerMemoria.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()

	utilsMemoria.InicializarMemoria()

	pid := 35
	utilsMemoria.InicializarMetricas(pid)
	utilsMemoria.CrearTablaPaginas(pid,217)

	fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid)
	utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")

	marco := utilsMemoria.EncontrarMarco(pid, []int{0, 0, 0, 0, 1}) // debería dar marco 1

	fmt.Printf("Marco encontrado a partir de entradas %d:\n", marco)

	utilsMemoria.ImprimirMetricas(pid)
	select{}
}

