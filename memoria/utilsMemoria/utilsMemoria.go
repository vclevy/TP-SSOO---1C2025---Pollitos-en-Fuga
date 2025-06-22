package utilsMemoria

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
"log"
	"github.com/sisoputnfrba/tp-golang/memoria/global"
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

//Tabla de paginas
var TablaDePaginasRaiz []*EntradaTP // una por proceso
var TablasPorProceso = make(map[int]*EntradaTP)

//Instrucciones de procesos
var instruccionesProcesos map[int]*[]string

//MERTRICAS
var metricas = make(map[int]*MetricasProceso)
type MetricasProceso struct { //son todas cantidades
	AcesosTP int
	InstruccionesSolicitadas int
	BajadasSWAP int
	SubidasMemoPpal int
	LecturasMemo int
	EscriturasMemo int
}

//SWAP
var SWAP = make(map[int]*Gorda)
type Gorda struct {
	Inicio int64
	Fin int64
}

func InicializarMemoria() {
	TamMemoria = global.ConfigMemoria.Memory_size
	TamPagina = global.ConfigMemoria.Page_Size
	CantNiveles = global.ConfigMemoria.Number_of_levels
	CantEntradas = global.ConfigMemoria.Entries_per_page
	Delay = global.ConfigMemoria.Memory_delay
	SwapPath = global.ConfigMemoria.Swapfile_path

	instruccionesProcesos = make(map[int]*[]string)

    MemoriaUsuario = make([]byte, TamMemoria)

	metricas = make(map[int]*MetricasProceso)

	SWAP = make(map[int]*Gorda)

	TablasPorProceso = make(map[int]*EntradaTP)

	//bitmap
    totalMarcos := TamMemoria / TamPagina
    MarcosLibres = make([]bool, totalMarcos)
    for i := range MarcosLibres {
        MarcosLibres[i] = true
    }
}

func InicializarMetricas (pid int) {
	if metricas[pid] == nil {
		metricas[pid] = &MetricasProceso{}
	}	
}

//---------------------------------------------------------------------CASOS DE USO---------------------------------------------------------------------
//INICIALIZAR PROCESO
func CargarProceso(pid int, ruta string) error {
	
	contenidoArchivo, err := os.ReadFile(ruta)
	if err != nil {
		log.Printf("Error leyendo pseudocódigo del PID %d en ruta '%s': %v", pid, ruta, err)
		return err
	}

	lineas := strings.Split(strings.TrimSpace(string(contenidoArchivo)), "\n")

	// Log de ayuda para debug
	log.Printf("PID %d - %d instrucciones cargadas desde '%s'\n", pid, len(lineas), ruta)

	instruccionesProcesos[pid] = &lineas
	return nil
}


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

//separar logica de crear tabla y asignar marcos
func AsignarMarcos(pid int) []int {
	return []int{}
}






//FINALIZAR PROCESO
func FinalizarProceso(pid int) string{
	//liberar memoria usuario
	marcosDelProceso := EncontrarMarcosDeProceso(pid)
	LiberarEspacioMemoria(pid, marcosDelProceso)

	//liberar instrucciones y borrar tabla de paginas
	delete(instruccionesProcesos, pid)
	delete(TablasPorProceso, pid)

	//devolver metricas
	m := metricas[pid]

	metricasLoggear := "## PID: " + strconv.Itoa(pid) + " - Proceso Destruido - " +
		"Métricas - " +
		"Acc.T.Pag: " + strconv.Itoa(m.AcesosTP) + "; " +
		"Inst.Sol.: " + strconv.Itoa(m.InstruccionesSolicitadas) + "; " +
		"SWAP: " + strconv.Itoa(m.BajadasSWAP) + "; " +
		"Mem.Prin.: " + strconv.Itoa(m.SubidasMemoPpal) + "; " +
		"Lec.Mem.: " + strconv.Itoa(m.LecturasMemo) + "; " +
		"Esc.Mem.: " + strconv.Itoa(m.EscriturasMemo) + ";"

		delete(metricas,pid)

		return metricasLoggear
}

//LECTURA
func LeerMemoria(pid int, direccionFisica int, tamanio int) []byte{
	time.Sleep(time.Millisecond * time.Duration(Delay)) 

	datos := MemoriaUsuario[direccionFisica : direccionFisica+tamanio] 
	//Lee desde dirFisica hasta dirfisica+tamanio
	metricas[pid].LecturasMemo++

	return datos

}

func LeerPaginaCompleta (pid int, direccionFisica int) []byte{ //Hace lo mismo que Devolver Lectura memoria, solo que el tamaño es el de la pagina
	time.Sleep(time.Millisecond * time.Duration(Delay))

	//offset := direccionFisica%TamPagina
	return LeerMemoria(pid, direccionFisica, TamPagina)
}

//ESCRITURA
func EscribirDatos(pid int, direccionFisica int, datos []byte) { 
	
	time.Sleep(time.Millisecond * time.Duration(Delay))
	//se para en la posicion pedida y escribe de ahi en adelante
    tamanioDatos := len(datos)

    // Validación de límites de memoria
    if direccionFisica+tamanioDatos > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

    copy(MemoriaUsuario[direccionFisica:], datos)
    metricas[pid].EscriturasMemo++
}

func ActualizarPaginaCompleta (pid int, direccionFisica int, datos []byte) {
	time.Sleep(time.Millisecond * time.Duration(Delay))

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


//OBTENER INSTRUCCIONES
func ObtenerInstruccion(pid int, pc int) (string, error) { //ESTO SIRVE PARA CPU
	instrucciones := ListaDeInstrucciones(pid)

	if pc < 0 || pc >= len(instrucciones) { //Si PC es menor a 0 o mayor al lista de instrucciones -> ERROR
		return "", fmt.Errorf("PC %d fuera de rango", pc)
	}
	metricas[pid].InstruccionesSolicitadas++
	return instrucciones[pc], nil
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






//SWAP
func Suspender(pid int) {
	marcosDelProceso := EncontrarMarcosDeProceso(pid)
	dataMarcos := EncontrarDataMarcos(marcosDelProceso)
	LiberarEspacioMemoria(pid, marcosDelProceso)
	GuardarInfoSwap(pid, dataMarcos)
}

func EncontrarDataMarcos(marcos []int) []byte {
	return []byte{}
}

func DesSuspenderProceso(pid int) {
	info := BuscarDataEnSwap(pid)
//	var marcosAsignados []int
	//marcosAginados := AsignarMarcos(pid)
	//idx := 0
	cantPaginas := len(info)/TamPagina
	for i:=0 ; i < cantPaginas; i++{
		//dirFisica := marcosAsignados[idx] * TamPagina
		
		//MemoriaUsuario.marcosAsignados[idx] = 
		//inicio := i * TamPagina
	//	fin := inicio + TamPagina
		//pagina := info[inicio:fin]

	//	copy(MemoriaUsuario[dirFisica:], pagina)
	}
}

func BuscarDataEnSwap(pid int) []byte{
	file, err := os.Open(SwapPath) //O_APPEND: Todo se agrega al final, no sobreescribe; O_CREATE: Si no existe lo crea; O_WORNLY Se abre solo para escritura, no lectura
	if err != nil{
		fmt.Printf("Error al abrir el archivo swap %v", err)
		return nil
	}
	defer file.Close()

	inicio := SWAP[pid].Inicio
	fin := SWAP[pid].Fin

	tamanio := fin - inicio
	buffer := make([]byte, tamanio) //buffer: tamaño de memoria temp

	file.ReadAt(buffer, inicio)

	return buffer
}

func GuardarInfoSwap(pid int, data []byte){

	//GPT ¿Como abrir archivo? 
	// 0644 => octal, dueño, grupo, otros => -rw-r--r--
	file, err := os.OpenFile(SwapPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //O_APPEND: Todo se agrega al final, no sobreescribe; O_CREATE: Si no existe lo crea; O_WORNLY Se abre solo para escritura, no lectura
	if err != nil{
		fmt.Printf("Error al abrir el archivo swap %v", err)
		return
	}
	defer file.Close()
	
	inicio, err := file.Seek(0, io.SeekEnd) //posicion del ultimo
	if err != nil{
		fmt.Printf("Error al ir a la ultima posición del archivo %v", err)
		return
	}

	file.Write(data)

	SWAP[pid].Inicio = inicio
	SWAP[pid].Fin = inicio + int64(len(data))

}


//OTROS
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

func LiberarEspacioMemoria(pid int, marcosALiberar []int) {
	for i := 0; i < len(marcosALiberar); i++ {
		idx := marcosALiberar[i]
		MarcosLibres[idx] = true
	}

}

func EncontrarMarcosDeProceso(pid int) []int {
	raiz := TablasPorProceso[pid]
	if raiz == nil {
		return nil
	}

	var marcos []int

	var recorrerTabla func(tabla []*EntradaTP)
	recorrerTabla = func(tabla []*EntradaTP) {
		for _, entrada := range tabla {
			if entrada == nil || !entrada.Presente {
				continue
			}
			if entrada.SiguienteNivel == nil {
				// Último nivel: tiene marco
				marcos = append(marcos, entrada.MarcoFisico)
			} else {
				// Nivel intermedio: seguir bajando
				recorrerTabla(entrada.SiguienteNivel)
			}
		}
	}

	// Comenzar desde la tabla raíz del proceso
	recorrerTabla(raiz.SiguienteNivel)
	return marcos
}



//VERIFICAR ESPACIO DISPONIBLE
func HayLugar(tamanio int)(bool){
	var cantMarcosLibres int
	for i := 0; i < len(MarcosLibres); i++ {
		if MarcosLibres[i] {
			cantMarcosLibres++
		}		
	}
	
	cantMarcosNecesitados:= int(math.Ceil(float64(tamanio) / float64(TamPagina)))
	return cantMarcosNecesitados <= cantMarcosLibres
}

//ver
func ArrayBytesToString(data []byte) string {
    // Primero lo convertís a string
    str := string(data)
    // Luego eliminás todos los espacios
    return strings.ReplaceAll(str, " ", "")
}

func ListaDeInstrucciones(pid int) ([]string) {
    return *instruccionesProcesos[pid]
}

//Datos del config
var TamMemoria int
var TamPagina int
var CantNiveles int
var CantEntradas int
var Delay int
var SwapPath string




//---------------------------------------------------------------------PRUEBAS---------------------------------------------------------------------
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

func FormatearMarcos(marcos []int) string {
	strs := make([]string, len(marcos))
	for i, m := range marcos {
		strs[i] = strconv.Itoa(m)
	}
	return strings.Join(strs, "-")
}
