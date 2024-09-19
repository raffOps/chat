package pkg

import "log"

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("\n%s: %s", msg, err)
	}
}
