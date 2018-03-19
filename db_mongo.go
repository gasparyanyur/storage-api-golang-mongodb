package main

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

type MongoConnection struct {
	Session *mgo.Session
	Gfs     *mgo.GridFS
}

var Connection *MongoConnection

func newMongoDB(addr string, cred *mgo.Credential) (*MongoConnection, error) {
	conn, err := mgo.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("mongo: could not dial: %v", err)
	}
	if cred != nil {
		if err := conn.Login(cred); err != nil {
			return nil, err
		}
	}

	return &MongoConnection{
		conn,
		conn.DB(config.dbName).GridFS("fs"),
	}, nil
}
