package logfile

import (
	"log"
	"os"
	"path/filepath"
)

var Log *log.Logger //log file

func init() { //log file open
	path := "log"
	CheckFolder(path)
	path = filepath.Join(path, "info.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600) // update path for your needs
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	Log = log.New(f, "", log.LstdFlags|log.Lshortfile)
}

func Println(M ...interface{}) {
	log.Println(M...)
	Log.Println(M...)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
	Log.Printf(format, v...)
}

func pathexists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckFolder(path string) {
	exist, err := pathexists(path)
	if err != nil {
		return
	}
	if exist {
	} else {
		os.Mkdir(path, os.ModePerm)
	}
}
