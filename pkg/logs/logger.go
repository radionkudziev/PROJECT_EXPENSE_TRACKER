package logs

import (
	"os"

	"github.com/labstack/gommon/log"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatal(args ...interface{})
}

func NewLogger(writeToFile bool, level string) *log.Logger {
	logger := log.New("expense-tracker")

	if writeToFile {
		// Создаем/открываем файл для записи логов
		logFile, err := os.OpenFile("logs/app.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0666)
		if err != nil {
			// Если не можем создать файл, логируем в консоль
			logger.Warnf("Failed to open log file: %v", err)
		} else {
			logger.SetOutput(logFile)
		}
	}

	// Устанавливаем уровень логирования
	switch level {
	case "debug":
		logger.SetLevel(log.DEBUG)
	case "warn":
		logger.SetLevel(log.WARN)
	case "error":
		logger.SetLevel(log.ERROR)
	default:
		logger.SetLevel(log.INFO)
	}

	// Настраиваем формат логов
	logger.SetHeader(`{"time":"${time_rfc3339}","level":"${level}","file":"${short_file}","line":"${line}","message":"${message}"}` + "\n")

	return logger
}

// Вспомогательная функция для создания логгера с дефолтными настройками
func NewDefaultLogger() *log.Logger {
	return NewLogger(false, "info")
}
