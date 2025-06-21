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

	//TESTING
	//probando creacion de TABLA DE PAGINAS y encontrar MARCO
	pid := 35
	tamanio := 400
	dirLogica := []int{0, 0, 3, 2, 1} //devuelve marco 57
	
	utilsMemoria.InicializarMetricas(pid)
	utilsMemoria.CrearTablaPaginas(pid,tamanio)

	fmt.Printf("Tablas de páginas del proceso PID %d:\n", pid)
	utilsMemoria.ImprimirTabla(utilsMemoria.TablasPorProceso[pid].SiguienteNivel, 1, "")

	marco := utilsMemoria.EncontrarMarco(pid, dirLogica)

	fmt.Printf("Marco encontrado a partir de entradas: %d\n", marco)

	utilsMemoria.ImprimirMetricas(pid)


	//-------------probando escritura y lectura memoria
	utilsMemoria.InicializarMetricas(pid)

	utilsMemoria.EscribirDatos(pid,256,"AXB3")

	lecturaIndividual := utilsMemoria.MemoriaUsuario[258]//deberia devolver B
	fmt.Println("Dato posicion 258: " + string(lecturaIndividual)) 
	
	lecturaCPU := utilsMemoria.DevolverLecturaMemoria(pid, 257, 3) //deberia devolver XB3
	fmt.Println("Datos leidos desde 257 + 3: " + lecturaCPU)

	lecturaPaginaCompleta := utilsMemoria.LeerPaginaCompleta(pid, 256)
	fmt.Println("Leyendo pagina completa a partir de 256: " + lecturaPaginaCompleta)

	utilsMemoria.ActualizarPaginaCompleta(pid, 256, "JH9")
	lecturaPaginaCompleta2 := utilsMemoria.LeerPaginaCompleta(pid, 256)
	fmt.Println("Leyendo pagina completa actualizada a partir de 256: " + lecturaPaginaCompleta2)
	
	//probando los casos en los que los datos son incorrectos
	lecturaIndividualVacia := utilsMemoria.MemoriaUsuario[267]//devuelve un string vacio
	fmt.Println("Dato posicion 267 vacia: " + string(lecturaIndividualVacia)) 
	
	lecturaPaginaCompletaDesalineada := utilsMemoria.LeerPaginaCompleta(pid, 258)
	fmt.Println("Leyendo pagina completa a partir de 258 que no es el principio: " + lecturaPaginaCompletaDesalineada)
	

	select{}
}

