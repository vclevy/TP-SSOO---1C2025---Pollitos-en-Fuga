package utilskernel

import(

	"github.com/sisoputnfrba/tp-golang/kernel/global"
)

func ObtenerDispositivoIO(nombreBuscado string) []*global.IODevice {
    var dispositivos []*global.IODevice
    for i := range global.IOConectados {
        if global.IOConectados[i].Nombre == nombreBuscado {
            dispositivos = append(dispositivos, &global.IOConectados[i])
        }
    }
    return dispositivos
}
