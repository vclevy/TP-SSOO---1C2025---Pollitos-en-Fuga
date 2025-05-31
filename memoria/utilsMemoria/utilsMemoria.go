package utilsMemoria

import (
	"fmt"
	"os"
	"strings"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
	"math"
)

var memoriaUsuario []byte
var marcosLibres []bool
var tamMemoria = global.ConfigMemoria.Memory_size
var tamPagina = global.ConfigMemoria.Page_Size
var cantNiveles = global.ConfigMemoria.Number_of_levels
var cantEntradas = global.ConfigMemoria.Entries_per_page


func InicializarMemoria() {
    tamMemoria := global.ConfigMemoria.Memory_size
    tamPagina := global.ConfigMemoria.Page_Size

	//la direccion fisica es un indice dentro del siguiente array
    var memoriaUsuario = make([]byte, tamMemoria)

    totalMarcos := tamMemoria / tamPagina
    var marcosLibres = make([]bool, totalMarcos)

    for i := range marcosLibres {
        marcosLibres[i] = true
    }
}

var diccionarioProcesosMemoria = make(map[int]*[]string ) //procesosMemoria crea un dicionario (mapa) de los procesos


type EntradaTP struct {
    Presente       bool
    MarcoFisico    int
    SiguienteNivel []*EntradaTP
}

var tablaDePaginasRaiz []*EntradaTP // una por proceso
var tablasPorProceso = map[int]*EntradaTP{} // clave: PID

var metricas = make(map[int]*MetricasProceso)

//se usan al momento de destruir un proceso
type MetricasProceso struct { //son todas cantidades
	AcesosTP int
	InstruccionesSolicitadas int
	BajadasSWAP int
	SubidasMemoPpal int
	LecturasMemo int
	EscriturasMemo int
}


func ListaDeInstrucciones(pid int) ([]string) {
    return *diccionarioProcesosMemoria[pid]
}


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

func espacioDisponible()(int){ //MOCKUP
	return 2048
}

func HayLugar(tamanio int)(bool){
	return tamanio<=espacioDisponible()
}


func traducirLogicaAFisica(pid int,direcionLogica int){

}

func traducirFiscaALogica(pid int, direcionLogica int){

}

func ReservarMarcos(pid int, tamanio int){
	cantPaginas := int(math.Ceil(float64(tamanio) / float64(tamPagina)))

		var asignados []int
		for i := 0; i < len(marcosLibres) && len(asignados) < cantPaginas; i++ {
			if marcosLibres[i] {
				marcosLibres[i] = false
				asignados = append(asignados, i)
			}
		}
}

func CrearTablaPaginas(pid int, tamanio int){

}
