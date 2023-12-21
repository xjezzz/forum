package forum_logger

import (
	"log"
	"os"
)

var InfoLog *log.Logger
var ErrorLog *log.Logger

func InitLoggers() {
	InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	ErrorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}
