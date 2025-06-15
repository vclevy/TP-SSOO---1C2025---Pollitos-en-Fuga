package utilsMemoria

import (
	"fmt"
	"os"
	"strings"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"math"
	"time"
)

//ESTRUCTURAS
//Memoria de usuario
var MemoriaUsuario []byte
var MarcosLibres []bool

type EntradaTP struct {
	Presente       bool          // indica si está en MP o SWAP (último nivel)
	MarcoFisico    int           // apunta al marco físico (último nivel)
	SiguienteNivel []*EntradaTP  // apunta a la subtabla (intermedios)
}

var TablaDePaginasRaiz []*EntradaTP // una por proceso
var TablasPorProceso = make(map[int]*EntradaTP)

//Diccionario de procesos
var diccionarioProcesosMemoria map[int]*[]string

//Datos del config
var TamMemoria int
var TamPagina int
var CantNiveles int
var CantEntradas int

//MERTRICAS
// Se usan al momento de destruir un proceso para el
var metricas = make(map[int]*MetricasProceso)


type MetricasProceso struct { //son todas cantidades
	AcesosTP int
	InstruccionesSolicitadas int
	BajadasSWAP int
	SubidasMemoPpal int
	LecturasMemo int
	EscriturasMemo int
}

func InicializarMemoria() {
	TamMemoria = global.ConfigMemoria.Memory_size
	TamPagina = global.ConfigMemoria.Page_Size
	CantNiveles = global.ConfigMemoria.Number_of_levels
	CantEntradas = global.ConfigMemoria.Entries_per_page

	//la direccion fisica es un indice dentro del siguiente array
	diccionarioProcesosMemoria = make(map[int]*[]string)

    MemoriaUsuario = make([]byte, TamMemoria)

	metricas = make(map[int]*MetricasProceso)

    totalMarcos := TamMemoria / TamPagina
    MarcosLibres = make([]bool, totalMarcos)

    for i := range MarcosLibres {
        MarcosLibres[i] = true
    }

	TablasPorProceso = make(map[int]*EntradaTP)
}

func InicializarMetricas (pid int) {
	if metricas[pid] == nil {
		metricas[pid] = &MetricasProceso{}
	}	
}

//DICCIONARIO DE PROCESOS
func CargarProceso(pid int, ruta string) error { 
	contenidoArchivo, err := os.ReadFile(ruta)
	if err != nil {
		return err
	}

	lineas := strings.Split(strings.TrimSpace(string(contenidoArchivo)), "\n") //Splitea las lineas del archivo segun un salto de linea y el TrimSpace elimina espacios en blanco (al principio y al final del archivo)

	diccionarioProcesosMemoria[pid] = &lineas

	return nil
}

func ObtenerInstruccion(pid int, pc int) (string, error) { //ESTO SIRVE PARA CPU
	instrucciones := ListaDeInstrucciones(pid)

	if pc < 0 || pc >= len(instrucciones) { //Si PC es menor a 0 o mayor al lista de instrucciones -> ERROR
		return "", fmt.Errorf("PC %d fuera de rango", pc)
	}
	metricas[pid].InstruccionesSolicitadas++
	return instrucciones[pc], nil
}

func ListaDeInstrucciones(pid int) ([]string) {
    return *diccionarioProcesosMemoria[pid]
}

//VERIFICAR ESPACIO DISPONIBLE
func HayLugar(tamanio int)(bool){
	var cantMarcosLibres int
	for i := 0; i < TamMemoria; i++ {
		if MarcosLibres[i] {
			cantMarcosLibres++
		}		
	}
	
	cantMarcosNecesitados:= int(math.Ceil(float64(tamanio) / float64(TamPagina)))
	return cantMarcosNecesitados <= cantMarcosLibres
}

//RESERVANDO MARCOS DEL PROCESO
func ReservarMarcos(tamanio int) []int{
		cantMarcos := int(math.Ceil(float64(tamanio) / float64(TamPagina)))
		var reservados []int
		for i := 0; i < len(MarcosLibres) && len(reservados) < cantMarcos; i++ {
			if MarcosLibres[i] {
				MarcosLibres[i] = false

				reservados = append(reservados, i)
			}
		}
		return reservados
}

//Cada índice de marco reservado (i) indica que los bytes de memoriaUsuario desde
// i*tamPagina hasta (i+1)*tamPagina-1 están asignados a un proceso -> CPU



func CrearTablaPaginas(pid int, tamanio int) {
	paginas := int(math.Ceil(float64(tamanio) / float64(TamPagina)))
	marcos := ReservarMarcos(tamanio) // slice con marcos reservados
	idx := 0                          // índice del próximo marco a asignar

	raiz := &EntradaTP{
		Presente:       false,
		SiguienteNivel: CrearTablaNiveles(1, &paginas, &marcos, &idx),
	}
	TablasPorProceso[pid] = raiz
}

func CrearTablaNiveles(nivelActual int, paginasRestantes *int, marcosReservados *[]int, proximoMarco *int) []*EntradaTP {
	tabla := make([]*EntradaTP, CantEntradas)

	// Inicializar todas las entradas con Presente = false
	for i := range tabla {
		tabla[i] = &EntradaTP{
			Presente:       false,
			MarcoFisico:    -1,
			SiguienteNivel: nil,
		}
	}

	for i := 0; i < CantEntradas && *paginasRestantes > 0; i++ {
		if nivelActual == CantNiveles {
			tabla[i].Presente = true
			tabla[i].MarcoFisico = (*marcosReservados)[*proximoMarco]
			*paginasRestantes--
			*proximoMarco++
		} else {
			subtabla := CrearTablaNiveles(nivelActual+1, paginasRestantes, marcosReservados, proximoMarco)
			if subtabla != nil {
				tabla[i].Presente = true
				tabla[i].SiguienteNivel = subtabla
			}
		}
	}

	return tabla
}

//ACCESO A TABLA DE PAGINAS
func EncontrarMarco(pid int, entradas []int) int {
	actual := TablasPorProceso[pid]
	
	if actual == nil {
		return -1 // error: no hay raíz
	}
	fmt.Printf("Entradas: %v (len = %d), cantNiveles = %d\n", entradas, len(entradas),  CantNiveles)

	for i := 0; i < len(entradas); i++ {

		idx := entradas[i]
		
		fmt.Printf("→ Nivel %d, idx %d, tabla len = %d\n", i+1, idx, len(actual.SiguienteNivel))
		
		//si esta fuera de rango
		if idx < 0 || idx >= len(actual.SiguienteNivel) {
			fmt.Printf("Nivel %d: índice %d fuera de rango (len = %d)\n", i+1, idx, len(actual.SiguienteNivel))
			return -1
		}
		
		//si estoy en un marco
		if actual.SiguienteNivel == nil {
			fmt.Printf("estoy en el ultimo nivel")
		}

		actual = actual.SiguienteNivel[idx]

		time.Sleep(time.Millisecond * time.Duration(global.ConfigMemoria.Memory_delay))

		metricas[pid].AcesosTP++
	}

	if !actual.Presente {
		return -1 // está en SWAP u otro error
	}

	return actual.MarcoFisico
}

//ACCESO A ESPACIO DE USUARIO
func DevolverLecturaMemoria(pid int, direccionFisica int, tamanio int) []byte{
	datos := MemoriaUsuario[direccionFisica : direccionFisica+tamanio] 
	//Lee desde dirFisica hasta dirfisica+tamanio
	metricas[pid].LecturasMemo++

	return datos //ver si tenemos q devolvr un array
}

func EscribirDatos(pid int, direccionFisica int, datos string) { //ACTUALIZADO 8-6-2025
	//se para en la posicion pedida y escribe de ahi en adelante
	bytesDatos := []byte(datos)
    tamanioDatos := len(bytesDatos)

    // Validación de límites de memoria
    if direccionFisica+tamanioDatos > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

    copy(MemoriaUsuario[direccionFisica:], bytesDatos)
    metricas[pid].EscriturasMemo++
}

func LeerPaginaCompleta (pid int, direccionFisica int) []byte{ //Hace lo mismo que Devolver Lectura memoria, solo que el tamaño es el de la pagina
	// el Byte 0 no es el index 0, sería el offset=0
	offset := direccionFisica%TamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina \n")
		return nil
	}
	return DevolverLecturaMemoria(pid, direccionFisica, TamPagina)
}

func ActualizarPaginaCompleta (pid int, direccionFisica int, datos []byte) {
	if len(datos) != TamPagina{
		fmt.Printf("Error: se esperaban %d bytes \n", TamPagina)
		return 
	}

	offset := direccionFisica%TamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina\n")
		return 
	}

	if direccionFisica+TamPagina > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

    copy(MemoriaUsuario[direccionFisica:direccionFisica+TamPagina], datos)
    metricas[pid].EscriturasMemo++
}

//SWAP

func GuardarInfoSwap(pid int){
	//no hay que guardar literalmente la tabla de paginas
	//habria q poner en la tabla de paginas q los marcos estan en swap P=0
	//habria que guardar todos los datos en el swap file bin
}
func LiberarEspacioMemoria(pid int) {
	//como encuentro los marcos que tiene asignado un proceso?
	//marcar como libre los marcos correspondientes
}


//----------PRUEBAS
func ImprimirTabla(tabla []*EntradaTP, nivel int, path string) {
	for i, entrada := range tabla {
		if entrada == nil {
			continue
		}
		prefijo := fmt.Sprintf("Nivel %d - Entrada %d (%s)", nivel, i, path)

		if entrada.SiguienteNivel == nil {
			fmt.Printf("%s → MARCO %d\n", prefijo, entrada.MarcoFisico)
		} else {
			ImprimirTabla(entrada.SiguienteNivel, nivel+1, fmt.Sprintf("%s->%d", path, i))
		}
	}
}

func ImprimirMetricas(pid int) {
	metrica := metricas[pid]

	if metrica == nil {
		fmt.Printf("No hay métricas para el PID %d\n", pid)
		return
	}

	fmt.Printf("Métricas para PID %d:\n", pid)
	fmt.Printf("  AccesosTP: %d\n", metrica.AcesosTP)
	fmt.Printf("  InstruccionesSolicitadas: %d\n", metrica.InstruccionesSolicitadas)
	fmt.Printf("  BajadasSWAP: %d\n", metrica.BajadasSWAP)
	fmt.Printf("  SubidasMemoPpal: %d\n", metrica.SubidasMemoPpal)
	fmt.Printf("  LecturasMemo: %d\n", metrica.LecturasMemo)
	fmt.Printf("  EscriturasMemo: %d\n", metrica.EscriturasMemo)
}
