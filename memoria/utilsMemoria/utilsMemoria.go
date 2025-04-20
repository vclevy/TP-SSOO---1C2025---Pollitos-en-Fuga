package utilsMemoria

import (
	"os"
	"strings"
)

//estructura PID -> Pseudocodigo: map[PID][]string
//Esto mapea un pid con un array de strings

type procesoMemoria struct{ //Ver fomato que necestian en KERNEL
	pid int
	instrucciones []string
	PC int //Program Counter
}

var diccionarioProcesosMemoria = make(map[int]*procesoMemoria) //procesosMemoria crea un dicionario (mapa) de los procesos

func cargarProceso(pid int, ruta string) error { 
	contenidoArchivo, err := os.ReadFile(ruta)
	if err != nil {
		return err
	}

	lineas := strings.Split(strings.TrimSpace(string(contenidoArchivo)), "\n") //Splitea las lineas del archivo segun un salto de linea y el TrimSpace elimina espacios en blanco (al principio y al final del archivo)

	diccionarioProcesosMemoria[pid] = &procesoMemoria{ //el identificador es el pid
		pid:           pid,
		instrucciones: lineas,
		PC:            0,
	}
	return nil
}