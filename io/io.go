package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	url := "http://localhost:8001/escribir" 
	body := []byte("hola desde IO")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Kernel:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de kernel:", resp.Status)
}
