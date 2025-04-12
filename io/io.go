package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	utils "github.com/sisoputnfrba/tp-golang/utils/config" 
	"github.com/sisoputnfrba/tp-golang/io/global"
)

func main() {

	global.IoConfig = utils.CargarConfig[global.Config]("config/config.json")
	puertoKernel := strconv.Itoa(global.IoConfig.Port_Kernel)
	url := "http://localhost:"+ puertoKernel +"/escribir" 
	body := []byte("hola desde IO")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Kernel:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de kernel:", resp.Status)
}
