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

utilsMemoria.InicializarMemoria()
	// configurar logger e inicializar config
	global.InitGlobal()
	defer global.LoggerMemoria.CloseLogger()

	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerMemoria.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()


	pid := 42
	tamanio := 768 // 768 / 64 = 12 páginas

	utilsMemoria.CrearTablaPaginas(pid, tamanio)

	fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid)
	utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")
	

	select{}
}

