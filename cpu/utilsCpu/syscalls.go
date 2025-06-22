package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	log "github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func Syscall_IO(instruccion Instruccion) {
	tiempo, err := strconv.Atoi(instruccion.Parametros[1])
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tiempo estimado: %v", log.ERROR)
		return
	}

	syscall_IO := estructuras.Syscall_IO{
		IoSolicitada:   instruccion.Parametros[0],
		TiempoEstimado: tiempo,
		PIDproceso:     global.PCB_Actual.PID,
	}

	jsonData, err := json.Marshal(syscall_IO)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/IO", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                      //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Init_Proc(instruccion Instruccion) {
	tamanio, err := strconv.Atoi(instruccion.Parametros[1]) //convieto tamanio de string a int
	if err != nil {
		global.LoggerCpu.Log("Error al convertir tamanio: %v", log.ERROR)
		return
	}

	syscall_Init_Proc := estructuras.Syscall_Init_Proc{
		ArchivoInstrucciones: instruccion.Parametros[0],
		Tamanio:              tamanio,
	}

	jsonData, err := json.Marshal(syscall_Init_Proc)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/Init_Proc", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))                             //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Dump_Memory() {
	url := fmt.Sprintf("http://%s:%d/Dump_Memory?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel, global.PCB_Actual.PID) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil)                                                                                   //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}

func Syscall_Exit() {
	url := fmt.Sprintf("http://%s:%d/exit?pid=%d", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel, global.PCB_Actual.PID) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", nil)                                                                            //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a Kernel: "+err.Error(), log.ERROR)
		return
	}

	defer resp.Body.Close() //se cierra la conexión
}
