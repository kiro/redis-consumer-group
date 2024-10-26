package internal

import (
	"log"
	"os"
)

var (
	INFO  = log.New(os.Stdout, "INFO", log.Llongfile|log.Ldate|log.Ltime|log.Lmsgprefix)
	ERROR = log.New(os.Stderr, "ERROR", log.Llongfile|log.Ldate|log.Ltime|log.Lmsgprefix)
)
