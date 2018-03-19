package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
)

func main() {
	log.Println("Starting server...")

	var err error
	var cred *mgo.Credential
	Connection, err = newMongoDB(config.dbHost, cred)
	if err != nil {
		log.Fatalf("mongo: could not dial: %v", err)
	}
	router := NewRouter()

	log.Fatal(http.ListenAndServe(config.serverHost, router))
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	} else {
		log.Printf(
			"%s %s - %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}

func appErrorfWithCode(err error, code int, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    code,
	}
}
