package utilsMemoria

import (
	"os"
	"strings"
	"fmt"
)

//estructura PID -> Pseudocodigo: map[PID][]string
//Esto mapea un pid con un array de strings

type procesoMemoria struct{ //Ver fomato que necestian en KERNEL
	pid int //sacar ver
	instrucciones []string 
	PC int //Program Counter
}

var diccionarioProcesosMemoria = make(map[int]*procesoMemoria) //procesosMemoria crea un dicionario (mapa) de los procesos

func CargarProceso(pid int, ruta string) error { 
	contenidoArchivo, err := os.ReadFile(ruta)
	if err != nil {
		return err
	}

	lineas := strings.Split(strings.TrimSpace(string(contenidoArchivo)), "\n") //Splitea las lineas del archivo segun un salto de linea y el TrimSpace elimina espacios en blanco (al principio y al final del archivo)

	diccionarioProcesosMemoria[pid] = &procesoMemoria{ //el identificador es el pid
		pid:           pid,
		instrucciones: lineas,
		PC:            0, //ver si esto se esta incrementando
	}
	return nil
}


func DevolverInstruccion(pid int, pc int) (string, error) { //ESTO SIRVE PARA CPU
	proceso, ok := diccionarioProcesosMemoria[pid]
	if !ok {
		return "", fmt.Errorf("PID %d no encontrado", pid)
	}
	if pc < 0 || pc >= len(proceso.instrucciones) { //Si PC es menor a 0 o mayor al lista de instrucciones -> ERROR
		return "", fmt.Errorf("PC %d fuera de rango", pc)
	}
	return proceso.instrucciones[pc], nil
}

func espacioDisponible()(int){ //MOCKUP
	return 2048
}

func HayLugar(tamanio int)(bool){
	return tamanio<=espacioDisponible()
}