package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	url := "http://localhost:8002/escribir" 
	body := []byte("hola desde kernel")
	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error al mandar mensaje a Memoria:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Respuesta de memoria:", resp.Status)
}
