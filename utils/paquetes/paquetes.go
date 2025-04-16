package paquetes

import (
	"bufio"
	"log"
	"os"
	"encoding/json"
	"bytes"
	"fmt"
	"net/http"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
}

func LeerConsola() Paquete {
	paquete := Paquete{}
	reader := bufio.NewReader(os.Stdin)
	log.Println("Ingrese los mensajes (Enter vacío para terminar):")

	for {
		text, _ := reader.ReadString('\n')

		if text == "\n" { // Enter vacío
			break
		}

		text = text[:len(text)-1] // Remueve el salto de línea
		log.Println("Mensaje ingresado:", text)
		paquete.Mensajes = append(paquete.Mensajes, text)
	}

	paquete.Codigo = 1 // Puedes cambiar el código según lo que necesites
	return paquete
}


func GenerarYEnviarPaquete(paquete Paquete, ip string, puerto int) {
	if len(paquete.Mensajes) == 0 {
		log.Println("No se ingresaron mensajes para enviar.")
		return
	}

	log.Printf("Paquete a enviar: %+v", paquete)
	EnviarPaquete(ip, puerto, paquete)
}

func EnviarPaquete(ip string, puerto int, paquete Paquete) {
	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensajes a ip:%s puerto:%d", ip, puerto)
	}

	log.Printf("respuesta del servidor: %s", resp.Status)
}
