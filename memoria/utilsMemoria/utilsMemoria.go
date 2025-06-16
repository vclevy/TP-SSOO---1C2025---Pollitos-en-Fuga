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

//Memoria de kernel
type EntradaTP struct {
    Presente       bool //indica si esta en MP o en SWAP (solo para el ultimo nivel)
    MarcoFisico    int //apunta al marco fisico de MP (solo para el ultimo nivel)
    SiguienteNivel []*EntradaTP 
}

var TablaDePaginasRaiz []*EntradaTP // una por proceso
var TablasPorProceso = make(map[int]*EntradaTP)

//Diccionario de procesos
var diccionarioProcesosMemoria map[int]*[]string

//Datos del config
var TamMemoria = global.ConfigMemoria.Memory_size
var tamPagina = global.ConfigMemoria.Page_Size
var cantNiveles = global.ConfigMemoria.Number_of_levels
var cantEntradas = global.ConfigMemoria.Entries_per_page


func InicializarMemoria() {
	//la direccion fisica es un indice dentro del siguiente array
	diccionarioProcesosMemoria = make(map[int]*[]string)

    MemoriaUsuario = make([]byte, TamMemoria)

    totalMarcos := TamMemoria / tamPagina
    MarcosLibres = make([]bool, totalMarcos)

    for i := range MarcosLibres {
        MarcosLibres[i] = true
    }

	TablasPorProceso = make(map[int]*EntradaTP)
}

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
	
	cantMarcosNecesitados:= int(math.Ceil(float64(tamanio) / float64(tamPagina)))
	return cantMarcosNecesitados <= cantMarcosLibres
}

//RESERVANDO MARCOS DEL PROCESO
func ReservarMarcos(tamanio int) []int{
		cantMarcos := int(math.Ceil(float64(tamanio) / float64(tamPagina)))
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

//CREACION DE TABLA DE PAGINAS DEL PROCESO
func CrearTablaPaginas(pid int, tamanio int){
	paginas := int(math.Ceil(float64(tamanio) / float64(tamPagina)))
	marcos := ReservarMarcos(tamanio) // slice con marcos reservados
	idx := 0 // índice para saber qué marco usar

	raiz := &EntradaTP{
		SiguienteNivel: CrearTablaNiveles(1, &paginas, &marcos, &idx),
	}
	TablasPorProceso[pid] = raiz
}

func CrearTablaNiveles(nivelActual int, paginasRestantes *int,marcosReservados *[]int,proximoMarco *int,) []*EntradaTP {
	tabla := make([]*EntradaTP, cantEntradas)

	for i := 0; i < cantEntradas; i++ {
		if nivelActual == cantNiveles {
			if *paginasRestantes > 0 {
				tabla[i] = &EntradaTP{
					Presente:       true,
					MarcoFisico:    (*marcosReservados)[*proximoMarco],
					SiguienteNivel: nil,
				}
				*paginasRestantes--
				*proximoMarco++
			} else {
				tabla[i] = nil
			}
		} else {
			tabla[i] = &EntradaTP{
				SiguienteNivel: CrearTablaNiveles(
					nivelActual+1,
					paginasRestantes,
					marcosReservados,
					proximoMarco,
				),
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

	for i := 0; i < len(entradas); i++ {
		idx := entradas[i]
		if actual.SiguienteNivel == nil || idx >= len(actual.SiguienteNivel) || idx < 0 {
			return -1 // error: tabla inválida
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
	offset := direccionFisica%tamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina \n")
		return nil
	}
	return DevolverLecturaMemoria(pid, direccionFisica, tamPagina)
}

func ActualizarPaginaCompleta (pid int, direccionFisica int, datos []byte) {
	if len(datos) != tamPagina{
		fmt.Printf("Error: se esperaban %d bytes \n", tamPagina)
		return 
	}

	offset := direccionFisica%tamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina\n")
		return 
	}

	if direccionFisica+tamPagina > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

    copy(MemoriaUsuario[direccionFisica:direccionFisica+tamPagina], datos)
    metricas[pid].EscriturasMemo++
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
