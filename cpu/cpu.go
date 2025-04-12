package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {

	//le habla a Memoria
	url := "http://localhost:8002/escribir" 
	body := []byte("hola desde CPU")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Memoria:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)

	//le habla a Kernel
	urlKernel := "http://localhost:8001/escribir" 
	bodyParaKernel := []byte("hola desde CPU")
	respKernel, errKernel := http.Post(urlKernel, "text/plain", bytes.NewBuffer(bodyParaKernel))
	if errKernel != nil {
		fmt.Println("Error al mandar mensaje a kernel:", errKernel)
		return
	}
	defer respKernel.Body.Close()

	fmt.Println("Respuesta de kernel:", respKernel.Status)


}