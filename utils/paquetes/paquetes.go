package paquetes


import (
	"bufio"
	"log"
	"os"
	"encoding/json"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type Paquete struct {
	Mensajes []string `json:"mensaje"`
	Codigo  	int    `json:"codigo"`
	PuertoDestino    int     `json:"puertoDestino"`
}

func LeerConsola() Paquete {
	paquete := Paquete{}
	reader := bufio.NewReader(os.Stdin)
	log.Println("Ingrese los mensajes (Enter vacío para terminar):")

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "" { // Enter vacío
			break
		}

		log.Println("Mensaje ingresado:", text)
		paquete.Mensajes = append(paquete.Mensajes, text)
	}

	// Solicitar el código al usuario
	log.Print("Ingrese el código del paquete: ")
	var codigo int
	for {
		_, err := fmt.Scanf("%d\n", &codigo)
		if err == nil {
			break
		}
		log.Print("Código inválido. Intente nuevamente: ")
		// Limpiar buffer en caso de error
		reader.ReadString('\n')
	}
	paquete.Codigo = codigo

	// Solicitar el puertoDestino al usuario
	log.Print("Ingrese el puerto destino del paquete: puerto CPU -> 8004, puerto Memoria -> 8002, puerto Kernel -> 8001, puerto Io -> 8003")
	var puerto int
	for {
		_, err := fmt.Scanf("%d\n", &puerto)
		if err == nil {
			break
		}
		log.Print("Puerto inválido. Intente nuevamente: ")
		// Limpiar buffer en caso de error
		reader.ReadString('\n')
	}
	paquete.PuertoDestino = puerto

	return paquete
}

func GenerarYEnviarPaquete(paquete Paquete, ip string) { //no se usaba el parametro de puerto porque directamente lo recibe del struct de paquete
	if len(paquete.Mensajes) == 0 {
		log.Println("No se ingresaron mensajes para enviar.")
		return
	}

	log.Printf("Paquete a enviar: %+v", paquete)
	EnviarPaquete(ip, paquete.PuertoDestino, paquete)
}

func EnviarPaquete(ip string, puerto int, paquete Paquete) {
	body, err := json.Marshal(paquete)
	if err != nil {
		log.Printf("error codificando mensajes: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/responder", ip, puerto)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("error enviando mensajes a ip:%s puerto:%d", ip, paquete.PuertoDestino)
	}

	log.Printf("Respuesta: %s", resp.Status)
}