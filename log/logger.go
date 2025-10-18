package log

import (
        "log"
        "os"
)

var writer = os.Stdout
var logger = log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)

const (
        prefixInfo = "[INFO] "
        prefixDebug = "[DEBUG] "
        prefixError = "[ERROR] "
        prefixWarn = "[WARN] "
)

func Info(format string, v ...any) {
        updFormat := prefixInfo + format
        logger.Printf(updFormat, v...)
}

func Debug(format string, v ...any) {
        updFormat := prefixDebug + format
        logger.Printf(updFormat, v...)
}

func Error(format string, v ...any) {
        updFormat := prefixError + format
        logger.Printf(updFormat, v...)
}

func Warn(format string, v ...any) {
        updFormat := prefixWarn + format
        logger.Printf(updFormat, v...)
}
