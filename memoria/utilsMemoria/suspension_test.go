package utilsMemoria

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/sisoputnfrba/tp-golang/memoria/global"
)

func TestFuncionamientoCompletoProceso(t *testing.T) {
	fmt.Println("▶️ Test: Funcionamiento completo - Suspensión y Des-suspensión")

	global.ConfigMemoria = &global.Config{
	Memory_size:       32,
	Page_Size:         4,
	Number_of_levels:  2,
	Entries_per_page:  4,
	Memory_delay:      0,
	Swapfile_path:     "test_swap_completo.bin",
	}


	InicializarMemoria()
	pid := 333
	tamanio := 12 // 3 páginas

	// Simular instrucciones
	instrucciones := []string{"LOAD", "ADD", "STORE"}
	instruccionesProcesos[pid] = &instrucciones

	// Crear tabla y escribir en memoria
	CrearTablaPaginas(pid, tamanio)
	marcos := EncontrarMarcosDeProceso(pid)
	if len(marcos) == 0 {
		t.Fatalf("❌ No se asignaron marcos al crear el proceso")
	}

	// Crear datos y escribirlos en memoria
	datosOriginales := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	for i := 0; i < len(marcos); i++ {
		inicio := marcos[i] * TamPagina
		fin := inicio + TamPagina
		copy(MemoriaUsuario[inicio:fin], datosOriginales[i*TamPagina:(i+1)*TamPagina])
	}

	// Suspender el proceso
	Suspender(pid)

	// Liberar la tabla y borrar memoria
	delete(TablasPorProceso, pid)

	// Des-suspender
	// Primero crear tabla vacía (sin marcos)
	CrearTablaPaginas(pid, tamanio)
	// Luego restaurar desde SWAP
	DesSuspenderProceso(pid)

	// Verificar que los datos restaurados coincidan
	marcosRestaurados := EncontrarMarcosDeProceso(pid)
	var datosRecuperados []byte
	for _, m := range marcosRestaurados {
		datosRecuperados = append(datosRecuperados, LeerPaginaCompleta(pid, m*TamPagina)...)
	}

	if !bytes.Equal(datosRecuperados, datosOriginales) {
		t.Errorf("❌ Falló la recuperación de datos.\n\tEsperado: %v\n\tObtenido: %v", datosOriginales, datosRecuperados)
	} else {
		t.Log("✅ Recuperación de datos exitosa.")
	}

	// Cleanup
	os.Remove(SwapPath)
}
