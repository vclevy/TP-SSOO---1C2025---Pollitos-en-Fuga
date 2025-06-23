package utilsMemoria

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

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
	MemoDelay = global.ConfigMemoria.Memory_delay
	SwapDelay = global.ConfigMemoria.Swap_delay
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

	global.MutexInstrucciones.Lock()
	instruccionesProcesos[pid] = &lineas
	global.MutexInstrucciones.Unlock()
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
func LeerMemoria(pid int, direccionFisica int, tamanio int) []byte {
	time.Sleep(time.Millisecond * time.Duration(MemoDelay)) 
	
	global.MutexMemoriaUsuario.Lock()
	datos := MemoriaUsuario[direccionFisica : direccionFisica+tamanio] 
	global.MutexMemoriaUsuario.Unlock()

	metricas[pid].LecturasMemo++

	return datos
}

func LeerPaginaCompleta (pid int, direccionFisica int) []byte{ //Hace lo mismo que Devolver Lectura memoria, solo que el tamaño es el de la pagina
	offset := direccionFisica%TamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina\n")
	}

	return LeerMemoria(pid, direccionFisica, TamPagina)
}

//ESCRITURA
func EscribirDatos(pid int, direccionFisica int, datos []byte) { 
	
	time.Sleep(time.Millisecond * time.Duration(MemoDelay))
	//se para en la posicion pedida y escribe de ahi en adelante
    tamanioDatos := len(datos)

    // Validación de límites de memoria
    if direccionFisica+tamanioDatos > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

	global.MutexMemoriaUsuario.Lock()
    copy(MemoriaUsuario[direccionFisica:], datos)
	global.MutexMemoriaUsuario.Unlock()

    metricas[pid].EscriturasMemo++
}

func ActualizarPaginaCompleta (pid int, direccionFisica int, datos []byte) {
	time.Sleep(time.Millisecond * time.Duration(MemoDelay))

	offset := direccionFisica%TamPagina
	if(offset!=0){
		fmt.Printf("Error: direccion física no alineada al byte 0 de la pagina\n")
		return 
	}

	if direccionFisica+TamPagina > len(MemoriaUsuario) {
        fmt.Printf("Error: intento de escritura fuera de los límites de memoria\n")
        return
    }

	global.MutexMemoriaUsuario.Lock()
    copy(MemoriaUsuario[direccionFisica:direccionFisica+TamPagina], datos)
	global.MutexMemoriaUsuario.Unlock()

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
	//fmt.Printf("Entradas: %v (len = %d), cantNiveles = %d\n", entradas, len(entradas),  CantNiveles)

	for i := 0; i < len(entradas); i++ {

		idx := entradas[i]
		
		//fmt.Printf("→ Nivel %d, idx %d, tabla len = %d\n", i+1, idx, len(actual.SiguienteNivel))
		
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
func SuspenderProceso(pid int) {
	time.Sleep(time.Millisecond * time.Duration(SwapDelay))
	marcosDelProceso := EncontrarMarcosDeProceso(pid)
	dataMarcos := EncontrarDataMarcos(marcosDelProceso)
	LiberarEspacioMemoria(pid, marcosDelProceso)
	GuardarInfoSwap(pid, dataMarcos)
}

func DesSuspenderProceso(pid int) {
	time.Sleep(time.Millisecond * time.Duration(SwapDelay))
	info := BuscarDataEnSwap(pid)
	tamanio := len(info) / TamPagina
	marcosAginados := ReservarMarcos(tamanio)
	fmt.Printf("🧩 Marcos asignados: %v (cantidad: %d)\n", marcosAginados, len(marcosAginados))//AUX
	fmt.Printf("📏 Info len: %d (espera %d marcos)\n", len(info), len(info)/TamPagina)//AUX
	AsignarMarcosATablaExistente(pid, marcosAginados)
	
	PegarInfoEnMemoria(pid, info, marcosAginados)
	fmt.Printf("📦 Info SWAP recuperada: len = %d, bytes = %v\n", len(info), info) //AUX

}

func EncontrarDataMarcos(marcos []int) []byte {
	var data []byte

	global.MutexMemoriaUsuario.Lock()
	for i := 0; i < len(marcos); i++ {
		inicio := marcos[i] * TamPagina
		fin := inicio + TamPagina

		// Validación, no creo que pase, pero ¿y si..?
		if inicio < 0 || fin > len(MemoriaUsuario) {
			fmt.Printf("Error: marco %d fuera de límites de memoria\n", marcos[i])
			continue //Saltea el data=append... y da una vuelta mas al for
		}

		data = append(data, MemoriaUsuario[inicio:fin]...)
	}
	global.MutexMemoriaUsuario.Unlock()

	return data
}

func PegarInfoEnMemoria(pid int, info []byte, marcosAsignados []int) {

	global.MutexMemoriaUsuario.Lock()
	for i := 0; i < len(marcosAsignados); i++ {
		inicio := marcosAsignados[i] * TamPagina
		fin := inicio + TamPagina

		copy(MemoriaUsuario[inicio:fin], info[i*TamPagina:(i+1)*TamPagina]) //copio
		fmt.Printf("✏️ Pegando página %d en marco %d [%d:%d]\n", i, marcosAsignados[i], i*TamPagina, (i+1)*TamPagina)//AUX

	}
	global.MutexMemoriaUsuario.Unlock()

	if metricas[pid] == nil {
		metricas[pid] = &MetricasProceso{}
	}
	metricas[pid].SubidasMemoPpal++
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

	global.MutexSwap.Lock()
	file.ReadAt(buffer, inicio)
	global.MutexSwap.Unlock()

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
	
	global.MutexSwap.Lock()
	inicio, err := file.Seek(0, io.SeekEnd) //posicion del ultimo
	if err != nil{
		fmt.Printf("Error al ir a la ultima posición del archivo %v", err)
		return
	}

	file.Write(data)
	global.MutexSwap.Unlock() //Si rompe, probar de poner debajo del lock con un defer... está acá para reducir la sección crítica

	if SWAP[pid] == nil {
		SWAP[pid] = &Gorda{}
	}

	SWAP[pid].Inicio = inicio
	SWAP[pid].Fin = inicio + int64(len(data))

}

//DUMP
func DumpMemoriaProceso(pid int){
	marcos := EncontrarMarcosDeProceso(pid)

	//CREAR ARCHIVO <PID>-<TIMESTAMP>.dmp
	timestamp := time.Now().Format("20060102-150405")
	nombreArchivo := fmt.Sprintf("%s/%d-%s.dmp", global.ConfigMemoria.Dump_path, pid, timestamp)

	file, err := os.Create(nombreArchivo) //Creo archivo
    if err != nil {
        log.Printf("❌ Error creando dump para PID %d: %v", pid, err)
        return
    }
    defer file.Close()

	for i := 0; i < len(marcos); i++ {
		inicio := marcos[i] * TamPagina
		fin := inicio + TamPagina

		datos:=MemoriaUsuario[inicio:fin]
		file.Write(datos)
	}
	log.Printf("✅ Dump de memoria creado: %s", nombreArchivo)
}


//OTROS
func ReservarMarcos(tamanio int) []int{
	cantMarcos := int(math.Ceil(float64(tamanio) / float64(TamPagina)))
	var reservados []int

	global.MutexMarcos.Lock()
	for i := 0; i < len(MarcosLibres) && len(reservados) < cantMarcos; i++ {
		if MarcosLibres[i] {
			MarcosLibres[i] = false

			reservados = append(reservados, i)
		}
	}
	global.MutexMarcos.Unlock()
	return reservados
}

func LiberarEspacioMemoria(pid int, marcosALiberar []int) {

	global.MutexMarcos.Lock()
	for i := 0; i < len(marcosALiberar); i++ {
		idx := marcosALiberar[i]
		MarcosLibres[idx] = true
	}
	global.MutexMarcos.Unlock()

	//habria que borrar la informacion que tienen
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

func AsignarMarcosATablaExistente(pid int, marcos []int) {
	tabla := TablasPorProceso[pid].SiguienteNivel
	proximo := 0
	recorrerYAsignar(tabla, &proximo, marcos, 1)
}

func recorrerYAsignar(tabla []*EntradaTP, proximo *int, marcos []int, nivelActual int) {
	for _, entrada := range tabla {
		if *proximo >= len(marcos) {
			return // ✅ Ya asignaste todos los marcos necesarios, corto el recorrido
		}

		if entrada == nil {
			continue
		}

		if nivelActual == CantNiveles {
			if *proximo < len(marcos) {
				fmt.Printf("📍 Asignando marco %d a entrada\n", marcos[*proximo]) //AUX
				entrada.MarcoFisico = marcos[*proximo]
				entrada.Presente = true
				*proximo += 1
			}
		} else if entrada.SiguienteNivel != nil {
			recorrerYAsignar(entrada.SiguienteNivel, proximo, marcos, nivelActual+1)
		}
		fmt.Printf("✅ Total marcos asignados: %d\n", proximo)//AUX
	}
}




//VERIFICAR ESPACIO DISPONIBLE
func HayLugar(tamanio int)(bool){
	var cantMarcosLibres int

	global.MutexMarcos.Lock()
	for i := 0; i < len(MarcosLibres); i++ {
		if MarcosLibres[i] {
			cantMarcosLibres++
		}		
	}
	global.MutexMarcos.Unlock()

	cantMarcosNecesitados:= int(math.Ceil(float64(tamanio) / float64(TamPagina)))
	return cantMarcosNecesitados <= cantMarcosLibres
}

func ListaDeInstrucciones(pid int) ([]string) {
    return *instruccionesProcesos[pid]
}

//Datos del config
var TamMemoria int
var TamPagina int
var CantNiveles int
var CantEntradas int
var MemoDelay int
var SwapDelay int
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
