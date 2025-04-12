package global

var CpuConfig *Config

type Config struct {
    IPMemory           	int 	`json:"ip_memory"`
    Port_Memory         int    	`json:"port_memory"`
	IPKernel 			string	`json:"ip_kernel"`
    Port_Kernel         int    	`json:"port_kernel"`
    PortCPU 		    int 	`json:"port_cpu"`
    TlbEntries     		int    	`json:"tlb_entries"`
    TlbReplacement      string 	`json:"tlb_replacement"`
	CacheEntries        int 	`json:"cache_entries"`
	CacheReplacement	int		`json:"cache_replacement"`
	CacheDelay			int		`json:"cache_delay"`
	LogLevel			string	`json:"log_level"`
	LogFile				string  `json:"log_file"`	
}	




// "port_cpu": 8004,
// "ip_memory": "127.0.0.1",
// "port_memory": 8002,
// "ip_kernel": "127.0.0.1",
// "port_kernel": 8002,
// "tlb_entries": 15,
// "tlb_replacement": "TLB",
// "cache_entries": 10,
// "cache_replacement": "CLOCK",
// "cache_delay": 10,
// "log_level": "DEBUG"