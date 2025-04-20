package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/io/global"
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"github.com/sisoputnfrba/tp-golang/utils/paquetes"
	//"github.com/sisoputnfrba/tp-golang/io/api"
)

func main() {

	global.InitGlobal() 
	defer global.LoggerIo.CloseLogger()

	//paso 1: leer como parametro el nombre de la interaz io desde consola
	//para chequear que se pase el nombre
	if len(os.Args) < 2 {
		global.LoggerIo.Log("Tenés que pasar el nombre de una interfaz. Ejemplo: go run ./src/io.go teclado", logger.ERROR)
        return
    }

    nombreInterfaz := os.Args[1]

    fmt.Printf("Se conectó %s!\n", nombreInterfaz)


	infoIO := paquetes.Paquete{
		Mensajes:      []string{"Nombre de IO:" + nombreInterfaz, global.IoConfig.IPIo ,strconv.Itoa(global.IoConfig.Port_Io)},
		Codigo:        200, //??????
		PuertoDestino: global.IoConfig.Port_Kernel,
	}
	
	respuesta, err := EnviarPaqueteAKernel(Paquete(infoIO), global.IoConfig.IPKernel)
	if err != nil {
	fmt.Println("Error al enviar paquete al Kernel:", err)
	return
}

fmt.Println("Respuesta del Kernel:")
fmt.Println("Status:", respuesta.Status)
fmt.Println("Detalle:", respuesta.Detalle)

	
	// s := api.CrearServer()
	// go func() {
	// 	err_server := s.Iniciar()
	// 	if err_server != nil {
	// 		global.LoggerIo .Log("Error al iniciar el servidor: "+err_server.Error(), log.ERROR)
	// 	}
	// }()
	
	//el kernel le va a mandar un paquete con el TIEMPO estimado que va a tardar y el PID asignado
	//api.RecibirPaquete() ..
	//global.LoggerIo.Log("##PID: "+ pidString + "- Inicio de IO - Tiempo " + tiempoIO, log.DEBUG) 
	// //Al momento de recibir una petición del Kernel, el módulo deberá iniciar un usleep por el tiempo indicado en la request
	// time.Sleep(time.Duration(tiempoIO) * time.Millisecond)
	// global.LoggerIo.Log("##PID: "+ pidString + "- FIn de IO", log.DEBUG)


}

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
}
type RespuestaKernel struct {
	Status         string `json:"status"`
	Detalle        string `json:"detalle"`
	PID            int    `json:"pid"`
	TiempoEstimado int    `json:"tiempo_estimado"`
}


func EnviarPaqueteAKernel(paquete Paquete, ip string) (*RespuestaKernel, error) {
	// Paso 1: Validar que haya mensajes en el paquete
	if len(paquete.Mensajes) == 0 {
		global.LoggerIo.Log("No se ingresaron mensajes para enviar.", "ERROR")
		return nil, fmt.Errorf("no se ingresaron mensajes para enviar")
	}

	// Paso 2: Log de paquete a enviar
	global.LoggerIo.Log(fmt.Sprintf("Paquete a enviar: %+v", paquete), "DEBUG")

	// Paso 3: Convertir el paquete a JSON
	body, err := json.Marshal(paquete)
	if err != nil {
		global.LoggerIo.Log("Error codificando paquete a JSON: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 4: Enviar el paquete al Kernel (POST)
	url := fmt.Sprintf("http://%s:%d/responder", ip, paquete.PuertoDestino)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		global.LoggerIo.Log(fmt.Sprintf("Error enviando paquete a %s:%d - %s", ip, paquete.PuertoDestino, err.Error()), "ERROR")
		return nil, err
	}
	defer resp.Body.Close()

	// Paso 5: Log de la respuesta HTTP
	global.LoggerIo.Log("Respuesta HTTP del Kernel: "+resp.Status, "DEBUG")

	// Paso 6: Leer y procesar la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.LoggerIo.Log("Error leyendo la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 7: Deserializar la respuesta
	var respuesta RespuestaKernel
	err = json.Unmarshal(respBody, &respuesta)
	if err != nil {
		global.LoggerIo.Log("Error parseando la respuesta del Kernel: "+err.Error(), "ERROR")
		return nil, err
	}

	// Paso 8: Loguear la respuesta del Kernel en el log de IO
	global.LoggerIo.Log(fmt.Sprintf("Respuesta del Kernel: Status=%s | Detalle=%s | PID=%d | TiempoEstimado=%dms",
		respuesta.Status, respuesta.Detalle, respuesta.PID, respuesta.TiempoEstimado), "DEBUG")

	// Devolver la respuesta
	return &respuesta, nil
}
