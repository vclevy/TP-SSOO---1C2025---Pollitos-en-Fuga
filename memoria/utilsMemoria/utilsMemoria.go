package utilsMemoria

import (
	"os"
	"strings"
	"fmt"
)

//estructura PID -> Pseudocodigo: map[PID][]string
//Esto mapea un pid con un array de strings


var diccionarioProcesosMemoria = make(map[int]*[]string ) //procesosMemoria crea un dicionario (mapa) de los procesos

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
	return instrucciones[pc], nil
}

func espacioDisponible()(int){ //MOCKUP
	return 2048
}

func HayLugar(tamanio int)(bool){
	return tamanio<=espacioDisponible()
}