package utilsIo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"github.com/sisoputnfrba/tp-golang/cpu/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	estructuras "github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"strings"
)

var instruccionesConMMU = map[string]bool{
	"WRITE":      true,
	"READ":       true,
	"GOTO":       true,
}

func Fetch(pid int, pc int) {

	solicitudInstruccion := estructuras.SolicitudInstruccion{
		Pid: pid,
		Pc:  pc,
	}

	jsonData, err := json.Marshal(solicitudInstruccion)
	if err != nil {
		global.LoggerCpu.Log("Error serializando solicitud: "+err.Error(), log.ERROR)
		return
	}

	url := fmt.Sprintf("http://%s:%d/solicitudInstruccion", global.CpuConfig.Ip_Memoria, global.CpuConfig.Port_Memoria) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión
	
	if err != nil {
		global.LoggerCpu.Log("Error enviando solicitud de instrucción a memoria: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Solicitud enviada a Memoria con éxito", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var instruccionAEjecutar string
	err = json.Unmarshal(body, &instruccionAEjecutar)
	if err != nil {
		global.LoggerCpu.Log("Error parseando instruccion de Memoria: "+err.Error(), log.ERROR)
		return
	}

	global.LoggerCpu.Log(fmt.Sprintf("Memoria respondió con la instrucción: %s", instruccionAEjecutar), log.INFO)

	Decode(instruccionAEjecutar)
}

type Instruccion struct {
	Opcode  string	`json:"opcode"`  // El tipo de operación (e.g. WRITE, READ, GOTO, etc.)
	Parametros []string `json:"parametros"` // Los parámetros de la instrucción, de tipo variable
}

func Decode(instruccionAEjecutar string){
	instruccionPartida := strings.Fields(instruccionAEjecutar)

	opcode := instruccionPartida[0]
	parametros := instruccionPartida[1:]

	instruccion := Instruccion{
		Opcode: opcode,
		Parametros:  parametros,
	}
	Execute(instruccion)
}

func Execute(instruccion Instruccion){
	if _, requiereMMU := instruccionesConMMU[instruccion.Opcode]; requiereMMU {
		MMU(instruccion)
	}
}

func MMU(instruccion Instruccion){
	
	/* traducir direcciones lógicas a físicas, 
		dirección logica [entrada_nivel_1 | entrada_nivel_2 | … | entrada_nivel_X | desplazamiento] 
		
		Teniendo una cantidad de niveles N y un identificador X de cada nivel podemos utilizar las siguientes fórmulas:
		nro_página = floor(dirección_lógica / tamaño_página)
		entrada_nivel_X = floor(nro_página  / cant_entradas_tabla ^ (N - X)) % cant_entradas_tabla
		desplazamiento = dirección_lógica % tamaño_página
	*/
}

func CheckInterrupt(instruccion Instruccion){}

func EnviarAKernel(){

}

/* 
TODO:
? usar query paths
? implementar que las funciones reciban errores(?) func Decode(instruccion string) (string, error) 
? hacer mmu
? delegar las syscalls a kernel, me devuelve algo kernel?
*/ 

/* 
func RealizarHandshakeConKernel() {
	type datosEnvio struct {
		Id		string 	 `json:"id"`
		Ip  	string   `json:"ip"`
		Puerto	int		 `json:"puerto"`
	}

	type datosRespuesta struct {
		Pid		int		`json:"Pid"`
		Pc		int		`json:"Pc"`
	}
 	
	//envio
	var envio datosEnvio

	jsonData, err := json.Marshal(envio)
	if err != nil {
		global.LoggerCpu.Log("Error serializando handshake: "+err.Error(), log.ERROR)
		return
	}
	
	url := fmt.Sprintf("http://%s:%d/handshake", global.CpuConfig.Ip_Kernel, global.CpuConfig.Port_Kernel) //url a la que se va a conectar
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData)) //se abre la conexión

	if err != nil {
		global.LoggerCpu.Log("Error enviando handshake al Kernel: " + err.Error(), log.ERROR)
		return
	}
	defer resp.Body.Close() //se cierra la conexión

	global.LoggerCpu.Log("✅ Handshake enviado al Kernel con éxito", log.INFO)

	//respuesta
	body, _ := io.ReadAll(resp.Body)

	var respuesta datosRespuesta
	err = json.Unmarshal(body, &respuesta)
	if err != nil {
		global.LoggerCpu.Log("Error parseando respuesta del Kernel: "+err.Error(), log.ERROR)
		return
	}

	global.LoggerCpu.Log(fmt.Sprintf("Kernel respondió con PID: %d y PC: %d", respuesta.Pid, respuesta.Pc), log.INFO)
}
*/

/* func SolicitarInstruccionAMemoria(pid int, pc int) {
	// Creamos la URL con los valores de pid y pc
	url := fmt.Sprintf("http://%s:%d/memoria/%d/%d",global.CpuConfig.Ip_Memoria,global.CpuConfig.Port_Memoria, pid, pc)


	// Realizamos el request GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error al crear la solicitud:", err)
		return
	}

	// Establecemos el tipo de contenido que estamos enviando
	req.Header.Set("Content-Type", "application/json")

	// Enviamos la solicitud
	cliente := &http.Client{}
	respuesta, err := cliente.Do(req)
	if err != nil {
		fmt.Println("Error al hacer la solicitud:", err)
		return
	}

	// Verificamos el código de estado
	if respuesta.StatusCode != http.StatusOK {
		fmt.Println("Error, estado de respuesta:", respuesta.Status)
		return
	}

	// Leemos el cuerpo de la respuesta
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}

	// Imprimimos la respuesta
	fmt.Println("Respuesta de Memoria:", string(bodyBytes))
} */