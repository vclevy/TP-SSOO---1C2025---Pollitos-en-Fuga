package global

import(
	logger "github.com/sisoputnfrba/tp-golang/utils/logger"
	"encoding/json"
	"os"
)
var MemoriaConfig Config

type Config struct {
	Port_Memory      int      `json:"port_memory"`
	Memory_size      int      `json:"memory_size"`
	Page_Size		 int  	  `json:"page_size"`
	Entries_per_page int      `json:"entries_per_page"`
	Number_of_levels int      `json:"number_of_levels"`
	Memory_delay     int      `json:"memory_delay"`
	Swapfile_path    string   `json:"swapfile_path"`
	Swap_delay 		 int      `json:"swap_delay"`
	Log_level        string   `json:"log_level"`
	Dump_path		 string   `json:"dump_path"`
	Log_file         string   `json:"log_file"`
}

var Logger *logger.LoggerStruct

func CargarConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic("No se pudo abrir el archivo de configuración: " + err.Error())
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&MemoriaConfig)
	if err != nil {
		panic("Error al parsear el archivo de configuración: " + err.Error())
	}
}