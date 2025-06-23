package utilsMemoria

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func TestSoloSuspension(t *testing.T) {
	fmt.Println("▶️ Test: Solo suspensión del proceso")

	global.ConfigMemoria = &global.Config{
		Memory_size:      64,
		Page_Size:        4,
		Number_of_levels: 2,
		Entries_per_page: 4,
		Memory_delay:     0,
		Swapfile_path:    "swap_test.bin",
	}

	InicializarMemoria()

	pid := 321
	CrearTablaPaginas(pid, 12)
	InicializarMetricas(pid)

	dato0 := []byte{100, 101, 102, 103}
	dato1 := []byte{110, 111, 112, 113}
	dato2 := []byte{120, 121, 122, 123}
	datosOriginales := append(append(dato0, dato1...), dato2...)

	marcos := EncontrarMarcosDeProceso(pid)
	if len(marcos) != 3 {
		t.Fatalf("❌ Se esperaban 3 marcos, pero se obtuvieron %d", len(marcos))
	}

	ActualizarPaginaCompleta(pid, marcos[0]*TamPagina, dato0)
	ActualizarPaginaCompleta(pid, marcos[1]*TamPagina, dato1)
	ActualizarPaginaCompleta(pid, marcos[2]*TamPagina, dato2)

	// Ejecutar suspensión
	Suspender(pid)

	// Leer lo que se escribió en SWAP
	datosSwap := BuscarDataEnSwap(pid)

	if !bytes.Equal(datosSwap, datosOriginales) {
		t.Errorf("❌ Datos en swap no coinciden.\nEsperado: %v\nObtenido: %v", datosOriginales, datosSwap)
	} else {
		fmt.Println("✅ Suspensión exitosa: los datos fueron correctamente guardados en SWAP.")
	}

	_ = os.Remove("swap_test.bin")
}
