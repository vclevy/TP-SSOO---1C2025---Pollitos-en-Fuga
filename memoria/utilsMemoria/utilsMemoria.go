package utilsMemoria

import (
	"fmt"
	"os"
	"strings"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"math"
)

//Memoria de usuario
var MemoriaUsuario []byte
var MarcosLibres []bool

//Memoria de kernel
type EntradaTP struct {
    Presente       bool //indica si esta en MP o en SWAP (solo para el ultimo nivel)
    MarcoFisico    int //apunta al marco fisico de MP (solo para el ultimo nivel)
    SiguienteNivel []*EntradaTP 
}

var tablaDePaginasRaiz []*EntradaTP // una por proceso
var tablasPorProceso = make(map[int]*EntradaTP)

//Diccionario de procesos
var diccionarioProcesosMemoria map[int]*[]string

func ListaDeInstrucciones(pid int) ([]string) {
    return *diccionarioProcesosMemoria[pid]
}

var tamMemoria = global.ConfigMemoria.Memory_size
var tamPagina = global.ConfigMemoria.Page_Size
var cantNiveles = global.ConfigMemoria.Number_of_levels
var cantEntradas = global.ConfigMemoria.Entries_per_page


func InicializarMemoria() {
	//la direccion fisica es un indice dentro del siguiente array
	diccionarioProcesosMemoria = make(map[int]*[]string)

    MemoriaUsuario = make([]byte, tamMemoria)

    totalMarcos := tamMemoria / tamPagina
    MarcosLibres = make([]bool, totalMarcos)

    for i := range MarcosLibres {
        MarcosLibres[i] = true
    }
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

//FUNCIONES UTILES
//carga las instrucciones de un proceso en el diccionario
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

//verifica que haya espacio disponible en MP o en SWAP??
func espacioDisponible()(int){ //MOCKUP
	return 2048
}

func HayLugar(tamanio int)(bool){
	return tamanio<=espacioDisponible()
}


func TraducirLogicaAFisica(pid int,direcionLogica int){
//sumar 1 a la metrica de acceso a memoria pr cada tabla recorrida
//considerar delay
}

func TraducirFiscaALogica(pid int, direcionLogica int){

}

func ReservarMarcos(pid int, tamanio int){
	cantMarcos := int(math.Ceil(float64(tamanio) / float64(tamPagina)))

		var asignados []int
		for i := 0; i < len(MarcosLibres) && len(asignados) < cantMarcos; i++ {
			if MarcosLibres[i] {
				MarcosLibres[i] = false
				asignados = append(asignados, i)
			}
		}
}

func CrearTablaPaginas(pid int, tamanio int){
	cantPaginas := int(math.Ceil(float64(tamanio) / float64(tamPagina)))
	paginas := cantPaginas
	raiz := &EntradaTP{
	SiguienteNivel: CrearTablaNiveles(1, cantNiveles, cantEntradas, &paginas),
}
tablasPorProceso[pid] = raiz

}


func CrearTablaNiveles(nivelActual int, maxNiveles int, cantEntradas int, paginasRestantes *int) []*EntradaTP {
	tabla := make([]*EntradaTP, cantEntradas)

	for i := 0; i < cantEntradas; i++ {
		if nivelActual == maxNiveles {
			if *paginasRestantes > 0 {
				tabla[i] = &EntradaTP{
					Presente:     false,
					MarcoFisico:  -1,
					SiguienteNivel: nil,
				}
				*paginasRestantes--
			} else {
				tabla[i] = nil // sin página asignada
			}
		} else {
			tabla[i] = &EntradaTP{
				SiguienteNivel: CrearTablaNiveles(nivelActual+1, maxNiveles, cantEntradas, paginasRestantes),
			}
		}
	}

	return tabla
}

func DevolverLecturaMemoria(pid int, direccionFisica int, tamanio int) []byte{
	datos := MemoriaUsuario[direccionFisica : direccionFisica+tamanio]
	metricas[pid].LecturasMemo++

	return datos
}