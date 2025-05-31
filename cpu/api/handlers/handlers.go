package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/global"
	utilsIo "github.com/sisoputnfrba/tp-golang/cpu/utilsCpu"
	"github.com/sisoputnfrba/tp-golang/utils/estructuras"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)