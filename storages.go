package main

import (
	"strconv"

	"errors"

	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StorageData struct {
	Id       bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Owner    int           `json:"owner,omitempty" bson:"owner,omitempty"`
	IsActive int32         `json:"active,omitempty" bson:"active,omitempty" `
	Size     int           `json:"size,omitempty" bson:"size,omitempty" `
	Used     int           `json:"used" bson:"used"`
}

func (s *StorageData) Create(data StorageData) (string, error) {
	id := bson.NewObjectId()
	data.Id = id

	count, err := Connection.Session.DB("cctv_storage").C("storages").Find(bson.M{"active": 1, "owner": data.Owner}).Count()

	if err != nil {
		return "", err
	}

	if count == 0 {
		data.IsActive = 1
	}

	err = Connection.Session.DB("cctv_storage").C("storages").Insert(data)

	if err != nil {
		return "", err
	}
	return id.Hex(), nil
}

func (s *StorageData) getFreeSpace(beingID string) int {
	current, err := s.GetActive(beingID)
	fmt.Println(err)

	if err != nil {
		return 0
	}

	return current.Size - current.Used
}

func (s *StorageData) Update(data StorageData, id bson.ObjectId) bool {

	err := Connection.Session.DB("cctv_storage").C("storages").UpdateId(id, bson.M{"$inc": data})
	if err != nil {
		return false
	}
	return true
}

func (s *StorageData) Clear(id string) error {
	var data []map[string]bson.ObjectId

	storage, err := strconv.Atoi(id)

	if err != nil {
		return err
	}

	err = Connection.Gfs.Find(bson.M{"metadata.sid": storage}).Select(bson.M{"_id": 1}).All(&data)

	if err != nil {
		return err
	}

	var ids []bson.ObjectId

	for _, v := range data {
		ids = append(ids, v["_id"])
		err := Connection.Gfs.RemoveId(v["_id"])

		if err != nil {
			return err
		}
	}

	_, err = Connection.Session.DB("cctv_storage").C("tree").RemoveAll(bson.M{"fid": bson.M{"$in": ids}})

	if err != nil {
		return err
	}

	_, err = Connection.Session.DB("cctv_storage").C("tree").UpdateAll(nil,
		bson.M{"$pull": bson.M{"ch": bson.M{"_id": bson.M{"$in": ids}}}})

	if err != nil {
		return err
	}

	_, err = Connection.Session.DB("cctv_storage").C("tree").UpdateAll(nil,
		bson.M{"$pull": bson.M{"pr": bson.M{"_id": bson.M{"$in": ids}}}})

	if err != nil {
		return err
	}

	err = Connection.Session.DB("cctv_storage").C("storages").Update(bson.M{"owner": id}, bson.M{"$set": bson.M{"used": 0}})

	if err != nil {
		return err
	}

	return nil
}

func (s *StorageData) Delete(id string) error {
	err := s.Clear(id)

	if err != nil {
		return err
	}

	err = Connection.Session.DB("cctv_storage").C("storages").RemoveId(id)

	if err != nil {
		return err
	}

	return nil
}

func (s *StorageData) GetSizeAndUsed(id bson.ObjectId) int {

	var info StorageData

	Connection.Session.DB("cctv_storage").C("storages").Find(bson.M{"_id": id}).Select(bson.M{"size": 1, "used": 1}).One(&info)

	return info.Size - info.Used
}

func (s *StorageData) GetActive(beingId string) (StorageData, error) {

	var data StorageData

	being, err := strconv.Atoi(beingId)

	if err != nil {
		return StorageData{}, err
	}

	err = Connection.Session.DB("cctv_storage").C("storages").Find(bson.M{"active": 1, "owner": being}).One(&data)

	if err != nil {
		return StorageData{}, err
	}

	return data, nil
}

func (s *StorageData) Activate(id, beingId string) error {

	current, err := s.GetActive(beingId)

	if err != nil {
		return err
	}

	sid := bson.ObjectIdHex(id)

	being, err := strconv.Atoi(beingId)

	if err != nil {
		return err
	}

	var nInfo map[string]int

	err = Connection.Session.DB("cctv_storage").C("storages").FindId(sid).Select(bson.M{"size": 1}).One(&nInfo)

	if err != nil {
		return err
	}

	if current.Used > nInfo["size"] {
		return errors.New("not enough free space")
	}

	count, err := Connection.Session.DB("cctv_storage").C("storages").Find(
		bson.M{"$or": []bson.M{
			{"_id": current.Id, "owner": being},
			{"_id": sid, "owner": being},
		}},
	).Count()

	if err != nil {
		return err
	}

	if count != 2 {
		return mgo.ErrNotFound
	}

	_, err = Connection.Session.DB("cctv_storage").C("fs.files").UpdateAll(bson.M{"metadata.sid": current.Id}, bson.M{"$set": bson.M{"metadata.sid": sid}})

	if err != nil {
		return err
	}

	err = Connection.Session.DB("cctv_storage").C("storages").UpdateId(current.Id, bson.M{"$set": bson.M{"active": 0, "used": 0}})

	if err != nil {
		return err
	}

	err = Connection.Session.DB("cctv_storage").C("storages").UpdateId(sid, bson.M{"$set": bson.M{"active": 1, "used": current.Used}})

	if err != nil {
		return err
	}

	return nil
}
