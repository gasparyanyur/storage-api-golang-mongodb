package main

import (
	"time"

	"encoding/json"
	"mime/multipart"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Camera struct {
	Id          bson.ObjectId  `json:"id" bson:"_id,omitempty"`
	ChunkSize   uint32         `json:"chunkSize" bson:"chunkSize"`
	ContentType string         `json:"contentType" bson:"contentType"`
	Filename    string         `json:"filename" bson:"filename"`
	Length      uint32         `json:"length" bson:"length"`
	Md5         string         `json:"md5" bson:"md5"`
	UploadDate  time.Time      `json:"uploadDate" bson:"uploadDate"`
	Metadata    CameraMetadata `json:"metadata" bson:"metadata"`
}

type CameraMetadata struct {
	CamId    string    `json:"cam_id" bson:"cam_id"`
	Date     time.Time `json:"date" bson:"date"`
	Duration uint64    `json:"duration" bson:"duration"`
}

func (c *Camera) SaveVideo(metadata []byte, head *multipart.FileHeader, buf []byte) (*Camera, error) {

	var meta CameraMetadata
	err := json.Unmarshal(metadata, &meta)
	if err != nil {
		return nil, err
	}

	ifExists, err := Connection.Session.DB("cctv_storage").GridFS("ts").Open(head.Filename)

	if ifExists == nil {
		mfile, err := Connection.Session.DB("cctv_storage").GridFS("ts").Create(head.Filename)
		if err != nil {
			return nil, err
		}

		mfile.SetMeta(meta)
		mfile.SetContentType(head.Header.Get("Content-Type"))

		_, err = mfile.Write(buf)
		if err != nil {
			return nil, err
		}

		err = mfile.Close()
		if err != nil {
			return nil, err
		}
		return &Camera{
			Id:          mfile.Id().(bson.ObjectId),
			Filename:    mfile.Name(),
			ContentType: mfile.ContentType(),
		}, nil
	}
	return nil, nil

}

func (c *Camera) Get(camId string, begin time.Time, end time.Time, checkstart bool, checkend bool) ([]string, error) {
	var files []Camera
	var filesName []string
	var query bson.M
	var err error
	if checkstart == false || begin.Unix() < 1500000000 || begin.Unix() > time.Now().Unix() || begin.Unix() > end.Unix() {
		query = bson.M{"metadata.cam_id": camId}
		err = Connection.Session.DB("cctv_storage").C("ts.files").Find(query).Sort("-uploadDate").Limit(3).All(&files)
	} else {
		if checkend == true && checkstart == true {
			query = bson.M{
				"metadata.cam_id": camId,
				"uploadDate": bson.M{
					"$gt": begin,
					"$lt": end,
				},
			}
		} else if checkstart == true && checkend == false {
			query = bson.M{
				"metadata.cam_id": camId,
				"uploadDate": bson.M{
					"$gt": begin,
				},
			}
		}

		err = Connection.Session.DB("cctv_storage").C("ts.files").Find(query).All(&files)
	}

	if err != nil {
		return nil, err
	}
	for _, f := range files {
		filesName = append(filesName, f.Filename)
	}
	return filesName, nil
}

func (c *Camera) FindTSFile(camId string, fileName string) (*mgo.GridFile, error) {
	var camera Camera

	err := Connection.Session.DB("cctv_storage").GridFS("ts").Find(bson.M{
		"metadata.cam_id": camId,
		"filename":        fileName,
	}).One(&camera)

	if err != nil {
		return nil, err
	}

	file, err := Connection.Session.DB("cctv_storage").GridFS("ts").OpenId(camera.Id)
	if err != nil {
		return nil, err
	}

	return file, nil

}
