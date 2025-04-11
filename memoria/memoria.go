package main

import (
    "fmt"
    "net/http"
	"io"
)

func main() {
    http.HandleFunc("/escribir", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Petición recibida en memoria")
        w.Write([]byte("Memoria escribió"))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error leyendo el cuerpo", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		fmt.Printf("Mensaje recibido de kernel: %s\n", string(body))
    })

    fmt.Println("Servidor de memoria corriendo en :8001")
    http.ListenAndServe(":8001", nil)
}
