package utilskernel

import(

	"github.com/sisoputnfrba/tp-golang/kernel/global"
)

func ObtenerDispositivoIO(nombreBuscado string) *global.IODevice {
    for _, io := range global.IOConectados {
        if io.Nombre == nombreBuscado {
            return &io
        }
    }
    return nil
}