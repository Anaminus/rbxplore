package main

import (
	"io/ioutil"
	"log"
	"os"
)

func InitDebug() {
	if Option.Debug {
		log.SetFlags(log.Ltime)
		log.SetOutput(os.Stdout)
	} else {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
}
