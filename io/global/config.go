// {
//     "ip_kernel": "127.0.0.1",
//     "port_kernel": 8001,
//     "port_io": 8003,
//     "log_level": "DEBUG"
//		"log_file": "io.log"
// }

package global


type Config struct {
	IPKernel           	string 	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    PortIo 		   		int 	`json:"port_io"`
	LogLevel			string	`json:"log_level"`
	LogFile				string  `json:"log_file"`	
	}	

var IoConfig *Config