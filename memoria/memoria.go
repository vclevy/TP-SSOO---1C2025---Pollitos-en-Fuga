package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/api"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"github.com/sisoputnfrba/tp-golang/memoria/utilsMemoria"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	//"github.com/sisoputnfrba/tp-golang/utils/paquetes"
	//"fmt"
)

//CONEXIÓN;
//[KERNEL] ➜ Cliente (conectado a) [MEMORIA]
//[CPU]    ➜ Cliente (conectado a) [MEMORIA]

func main() {

	global.InitGlobal()
	defer global.LoggerMemoria.CloseLogger()	
	
	utilsMemoria.InicializarMemoria()
	// configurar logger e inicializar config
	s := api.CrearServer()
	go func() {
		err_server := s.Iniciar()
		if err_server != nil {
			global.LoggerMemoria.Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
		}
		}()

	//TESTING
	//probando creacion de TABLA DE PAGINAS y encontrar MARCO
	// pid := 35
	// tamanio := 400
	// dirLogica := []int{0, 0, 3, 2, 1} //devuelve marco 57
	
	// utilsMemoria.InicializarMetricas(pid)
	// utilsMemoria.CrearTablaPaginas(pid,tamanio)

	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")

	// marco := utilsMemoria.EncontrarMarco(pid, dirLogica)

	// fmt.Printf("Marco encontrado a partir de entradas: %d\n", marco)

	// utilsMemoria.ImprimirMetricas(pid)


	// //-------------probando escritura y lectura memoria
	// utilsMemoria.InicializarMetricas(pid)

	// utilsMemoria.EscribirDatos(pid,256,"AXB3")

	// lecturaIndividual := utilsMemoria.MemoriaUsuario[258]//deberia devolver B
	// fmt.Println("Dato posicion 258: " + string(lecturaIndividual)) 
	
	// lecturaCPU := utilsMemoria.LeerMemoria(pid, 257, 3) //deberia devolver XB3
	// fmt.Println("Datos leidos desde 257 + 3: " + lecturaCPU)

	// lecturaPaginaCompleta := utilsMemoria.LeerPaginaCompleta(pid, 256)
	// fmt.Println("Leyendo pagina completa a partir de 256: " + lecturaPaginaCompleta)

	// utilsMemoria.ActualizarPaginaCompleta(pid, 256, "JH9")
	// lecturaPaginaCompleta2 := utilsMemoria.LeerPaginaCompleta(pid, 256)
	// fmt.Println("Leyendo pagina completa actualizada a partir de 256: " + lecturaPaginaCompleta2)
	
	// //probando los casos en los que los datos son incorrectos
	// lecturaIndividualVacia := utilsMemoria.MemoriaUsuario[267]//devuelve un string vacio
	// fmt.Println("Dato posicion 267 vacia: " + string(lecturaIndividualVacia)) 
	
	// lecturaPaginaCompletaDesalineada := utilsMemoria.LeerPaginaCompleta(pid, 258)
	// fmt.Println("Leyendo pagina completa a partir de 258 que no es el principio: " + lecturaPaginaCompletaDesalineada)
	
	//----------creando mas de un proceso
	// pid1 := 1
	// tamanio1 := 500

	// pid2 := 2
	// tamanio2 := 420

	// pid3 := 3
	// tamanio3 := 857

	// pid4 := 4
	// tamanio4 := 1050

	// utilsMemoria.InicializarMetricas(pid1)
	// utilsMemoria.InicializarMetricas(pid2)
	// utilsMemoria.InicializarMetricas(pid3)
	// utilsMemoria.InicializarMetricas(pid4)

	// utilsMemoria.CrearTablaPaginas(pid1,tamanio1)
	// utilsMemoria.CrearTablaPaginas(pid2,tamanio2)
	// utilsMemoria.CrearTablaPaginas(pid3,tamanio3)


	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid1)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid1].SiguienteNivel, 1, "")

	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid2)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid2].SiguienteNivel, 1, "")

	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid3)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid3].SiguienteNivel, 1, "")

	// marcos := utilsMemoria.EncontrarMarcosDeProceso(pid1)
	// utilsMemoria.FormatearMarcos(marcos)	
	// fmt.Printf("Marcos asignados %d:\n", marcos)
	
	// fmt.Println("--------------finalizo uno")
	// utilsMemoria.FinalizarProceso(pid2)
	// utilsMemoria.CrearTablaPaginas(pid4,tamanio4)

	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid4)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid4].SiguienteNivel, 1, "")
	
	//---------probando SWAP
	// pid := 1
	// tamanio := 500

	// utilsMemoria.InicializarMetricas(pid)
	// utilsMemoria.CrearTablaPaginas(pid,tamanio)

	// fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")

	// utilsMemoria.AsignarMarcosATablaExistente(pid, []int{6,9,0,4,12,3,8,1})
	// fmt.Printf("Tablas de páginas del proceso PID %d:\n CON NUEVOS MARCOS", pid)
	// utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")

	select{}
}

