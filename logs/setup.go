package log

import (
	"fmt"
	l "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
	pref "github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

var std *logrus.Logger

func init() {
	std = logrus.New()

	std.SetOutput(os.Stdout)

	std.SetFormatter(&pref.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})

	std.SetLevel(logrus.TraceLevel)
}

func Setup(serviceName string) {
	std.SetFormatter(&LogstashMessageFormatter{})

	//pathDev := "C:\\Users\\kompik\\GolandProjects\\awesomeProject\\" + serviceName + ".log"
	prod := "/root/micro/microParser/" + serviceName + ".log"

	file, err := os.OpenFile(
		prod,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		fmt.Println(err)

		std.Fatal(err)
	}

	std.Formatter = l.DefaultFormatter(logrus.Fields{"service": serviceName})
	std.SetOutput(file)

	std.SetLevel(logrus.InfoLevel)
}

func Info(args ...interface{}) {
	std.Info(args)
}

func Error(args ...interface{}) {
	std.Error(args)
}

func Fatal(args ...interface{}) {
	std.Fatal(args)
}
