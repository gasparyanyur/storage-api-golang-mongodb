package main

import (
	"encoding/json"
	"net/http"

	"io/ioutil"

	"strconv"
	"time"

	"regexp"

	"os"

	"errors"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	f File
	c Camera
	s StorageData
)

func Index(w http.ResponseWriter, r *http.Request) *appError {
	w.Write([]byte(`{
    "_id" : ObjectId("59c0e9896b9600db53597a67"),
    "chunkSize" : NumberInt(261120),
    "uploadDate" : ISODate("2017-09-19T09:55:21.821+0000"),
    "length" : NumberInt(0),
    "md5" : "d41d8cd98f00b204e9800998ecf8427e",
    "filename" : "1",
    "contentType" : "application/vnd.dantser-apps.folder",
    "metadata" : {
        "oid" : NumberInt(123),
        "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
        "kind" : "drive#folder",
        "parents" : [

        ]
    },
    "p" : [
        {
            "_id" : ObjectId("59c0e99a6b9600db53597a6b"),
            "chunkSize" : NumberInt(261120),
            "uploadDate" : ISODate("2017-09-19T09:55:38.490+0000"),
            "length" : NumberInt(0),
            "md5" : "d41d8cd98f00b204e9800998ecf8427e",
            "filename" : "1",
            "contentType" : "application/vnd.dantser-apps.folder",
            "metadata" : {
                "oid" : NumberInt(123),
                "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
                "kind" : "drive#folder",
                "parents" : [
                    ObjectId("59c0e9966b9600db53597a69")
                ]
            }
        },
        {
            "_id" : ObjectId("59c0e9966b9600db53597a69"),
            "chunkSize" : NumberInt(261120),
            "uploadDate" : ISODate("2017-09-19T09:55:34.903+0000"),
            "length" : NumberInt(0),
            "md5" : "d41d8cd98f00b204e9800998ecf8427e",
            "filename" : "1",
            "contentType" : "application/vnd.dantser-apps.folder",
            "metadata" : {
                "oid" : NumberInt(123),
                "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
                "kind" : "drive#folder",
                "parents" : [
                    ObjectId("59c0e9896b9600db53597a67")
                ]
            }
        }
    ]
}
{
    "_id" : ObjectId("59c0e9966b9600db53597a69"),
    "chunkSize" : NumberInt(261120),
    "uploadDate" : ISODate("2017-09-19T09:55:34.903+0000"),
    "length" : NumberInt(0),
    "md5" : "d41d8cd98f00b204e9800998ecf8427e",
    "filename" : "1",
    "contentType" : "application/vnd.dantser-apps.folder",
    "metadata" : {
        "oid" : NumberInt(123),
        "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
        "kind" : "drive#folder",
        "parents" : [
            ObjectId("59c0e9896b9600db53597a67")
        ]
    },
    "p" : [
        {
            "_id" : ObjectId("59c0e99a6b9600db53597a6b"),
            "chunkSize" : NumberInt(261120),
            "uploadDate" : ISODate("2017-09-19T09:55:38.490+0000"),
            "length" : NumberInt(0),
            "md5" : "d41d8cd98f00b204e9800998ecf8427e",
            "filename" : "1",
            "contentType" : "application/vnd.dantser-apps.folder",
            "metadata" : {
                "oid" : NumberInt(123),
                "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
                "kind" : "drive#folder",
                "parents" : [
                    ObjectId("59c0e9966b9600db53597a69")
                ]
            }
        }
    ]
}
{
    "_id" : ObjectId("59c0e99a6b9600db53597a6b"),
    "chunkSize" : NumberInt(261120),
    "uploadDate" : ISODate("2017-09-19T09:55:38.490+0000"),
    "length" : NumberInt(0),
    "md5" : "d41d8cd98f00b204e9800998ecf8427e",
    "filename" : "1",
    "contentType" : "application/vnd.dantser-apps.folder",
    "metadata" : {
        "oid" : NumberInt(123),
        "sid" : ObjectId("59c0de9a6b9600dad4ed0d3c"),
        "kind" : "drive#folder",
        "parents" : [
            ObjectId("59c0e9966b9600db53597a69")
        ]
    },
    "p" : [

    ]
}
`))
	return nil
}

func FileList(w http.ResponseWriter, r *http.Request) *appError {

	res, err := f.List()
	if err != nil {
		return appErrorf(err, "%+v", err)
	}
	r1, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(r1)
	return nil
}

func FileGet(w http.ResponseWriter, r *http.Request) *appError {
	var vars = mux.Vars(r)
	var fileId = vars["fileId"]

	res, err := f.Get(fileId)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}
	r1, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(r1)
	return nil
}

func GetDeleted(w http.ResponseWriter, r *http.Request) *appError {

	beingId, exists := r.Header["X-Being-Id"]

	if exists != true {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	being, err := strconv.Atoi(beingId[0])

	if err != nil {
		return appErrorf(err, "%+v")
	}

	files, err := f.GetDeleted(being)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	resp, err := json.Marshal(files)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(resp)

	return nil
}

func TextFileInfo(w http.ResponseWriter, r *http.Request) *appError {
	fileID := mux.Vars(r)["fileID"]
	var tf TextFile
	err := Connection.Session.DB("cctv_storage").C("text_files").FindId(bson.ObjectIdHex(fileID)).One(&tf)

	if err != nil {
		return appErrorf(err, "%+v")
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write(tf.Content)
	return nil
}

func UploadTextFile(w http.ResponseWriter, r *http.Request) *appError {

	folderID := mux.Vars(r)["folderID"]

	var fm Metadata
	var tf TextFile

	data, err := Connection.Gfs.OpenId(bson.ObjectIdHex(folderID))

	if err != nil {
		return appErrorf(err, "%+v")
	}

	data.GetMeta(&fm)

	if fm.Kind != "drive#folder" {
		return nil
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	json.Unmarshal(body, &tf)

	id := f.CreateTextFile(tf)

	if err != nil {
		return appErrorf(err, "%+v")
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"id":` + id + `}`))
	return nil
}

func UpdateTextFile(w http.ResponseWriter, r *http.Request) *appError {

	fileID := mux.Vars(r)["fileID"]

	var data TextFile

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	json.Unmarshal(body, &data)

	data.Content = []byte(data.Text)

	data.Text = ""

	updated := f.UpdateTexFile(bson.ObjectIdHex(fileID), data)

	var code int = 200

	if updated != true {
		code = 400
	}

	w.WriteHeader(code)

	return nil
}

func SimpleUpload(w http.ResponseWriter, r *http.Request) *appError {

	beingID := r.Header["X-Being-Id"][0]

	// todo Optimize limit read with Content-Length
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	var sd StorageData

	free := sd.getFreeSpace(beingID)

	if len(buf) > free {
		w.WriteHeader(409)
		return nil
	}

	file, err := f.SimpleCreate(buf, beingID)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(file)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	w.Write(res)

	return nil
}

func GetTree(w http.ResponseWriter, r *http.Request) *appError {

	id := mux.Vars(r)["fileID"]

	data, err := f.Childs(bson.ObjectIdHex(id))

	if err != nil {
		return appErrorf(err, "%+v")
	}

	response, err := bson.Marshal(data)

	if err != nil {
		return appErrorf(err, "%+v")
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(response)
	return nil
}

func MultiPartUpload(w http.ResponseWriter, r *http.Request) *appError {

	beingID := r.Header["X-Being-Id"][0]

	r.ParseMultipartForm(0)

	form := r.MultipartForm

	metadata := form.Value["meta"]

	var fm Metadata

	mt_err := json.Unmarshal([]byte(metadata[0]), &fm)

	if mt_err != nil {
		return appErrorf(mt_err, "%+v")
	}

	being, err := strconv.Atoi(beingID)

	fm.OwnerId = being

	mtype := r.Header["X-Upload-Content-Type"][0]

	var mfile *File

	if mtype == "application/vnd.dantser-apps.folder" {

		var err error
		mfile, err = f.CreateFolder(fm, beingID)

		if err != nil {
			return appErrorf(err, "%+v")
		}
		response := `{"id":"` + mfile.Id.Hex() + `"}`

		w.WriteHeader(201)
		w.Write([]byte(response))
		return nil

	} else {
		file := form.File["file"]

		cnt, err := file[0].Open()

		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		buf, err := ioutil.ReadAll(cnt)

		var sd StorageData

		free := sd.getFreeSpace(beingID)

		if len(buf) > free {
			w.WriteHeader(409)
			return nil
		}

		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		mfile, err = f.MultiPartCreate(buf, fm, beingID)

		if err != nil {
			return appErrorf(err, "%+v", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(mfile)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	w.Write(res)

	return nil
}

func ResumableUploadFile(w http.ResponseWriter, r *http.Request) *appError {
	beingID := r.Header["X-Being-Id"][0]

	fileID := mux.Vars(r)["fileID"]
	body, err := ioutil.ReadAll(r.Body)

	var (
		close bool
		fm    Metadata
	)

	if fileID == "" {
		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		err = json.Unmarshal(body, &fm)

		id, err := f.ResumableServeMetadata(fm, nil)

		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		w.Header().Set("Location", "http://"+r.Host+r.URL.Path+"/"+id+"?uploadType=resumable")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		return appErrorfWithCode(nil, 200, "")
	} else {
		uploadContentType := r.Header["X-Upload-Content-Type"][0]

		uploadLength, err := strconv.Atoi(r.Header["X-Upload-Content-Length"][0])

		var sd StorageData

		free := sd.getFreeSpace(beingID)

		if uploadLength > free {
			w.WriteHeader(409)
			return nil
		}

		contentRange := r.Header["Content-Range"][0]

		close = checkChunkIndex(contentRange)

		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		code, err := f.ResumableServeChunkedFile(body, uploadContentType, uploadLength, close)

		if err != nil {
			return appErrorfWithCode(err, 400, "")
		}
		sd.Used = uploadLength

		active, err := sd.GetActive(beingID)

		sd.Update(sd, active.Id)

		w.WriteHeader(code)

		return appErrorfWithCode(nil, code, "")
	}
}

func UpdateFile(w http.ResponseWriter, r *http.Request) *appError {
	fileid := mux.Vars(r)["fileID"]

	var insertData = make(map[string]interface{})

	if bson.IsObjectIdHex(fileid) {
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			appErrorf(err, "%+v")
		}

		var dt map[string]interface{}

		json.Unmarshal([]byte(body), &dt)

		insertData["$set"] = dt

		err = f.Update(bson.ObjectIdHex(fileid), insertData)

		if err != nil {
			return appErrorf(err, "%+v")
		}

		return nil

	} else {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	return nil
}

func CopyFile(w http.ResponseWriter, r *http.Request) *appError {
	fileid := mux.Vars(r)["fileID"]
	if bson.IsObjectIdHex(fileid) {
		id := bson.ObjectIdHex(fileid)
		err := f.Copy(id)

		if err != nil {
			return appErrorf(err, "%+v")
		}
		return appErrorfWithCode(nil, 201, "", "")
	}

	return appErrorf(mgo.ErrNotFound, "%+v")
}

func Delete(w http.ResponseWriter, r *http.Request) *appError {

	id := mux.Vars(r)["Id"]

	beingId := r.Header["X-Being-Id"][0]

	being, _ := strconv.Atoi(beingId)

	count, err := Connection.Gfs.Find(bson.M{"_id": bson.ObjectIdHex(id), "metadata.oid": being}).Count()

	if err != nil {
		return appErrorf(err, "%+v")
	}

	if count == 0 {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	err = f.Delete(bson.ObjectIdHex(id))

	if err != nil {
		return appErrorf(err, "%+v")
	}

	return nil

}

func ForceDelete(w http.ResponseWriter, r *http.Request) *appError {
	id := mux.Vars(r)["Id"]

	beingId, exists := r.Header["X-Being-Id"]

	if exists == false {
		appErrorf(errors.New("Being Id is required"), "%+v")
	}

	being, _ := strconv.Atoi(beingId[0])

	count, err := Connection.Gfs.Find(bson.M{"_id": bson.ObjectIdHex(id), "metadata.oid": being}).Count()

	if count == 0 || err != nil {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	err = f.ForceDelete(bson.ObjectIdHex(id))

	if err != nil {
		return appErrorf(err, "%+v")
	}

	w.WriteHeader(200)
	return nil

}

func Restore(w http.ResponseWriter, r *http.Request) *appError {

	id := mux.Vars(r)["fileID"]

	beingId := r.Header["X-Being-Id"][0]

	being, _ := strconv.Atoi(beingId)

	count, err := Connection.Gfs.Find(bson.M{"_id": bson.ObjectIdHex(id), "metadata.oid": being}).Count()

	if count == 0 || err != nil {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	err = f.Restore(bson.ObjectIdHex(id))

	if err != nil {
		return appErrorf(err, "%+v")
	}
	w.WriteHeader(200)
	return nil
}

func BulkDownloadFile(w http.ResponseWriter, r *http.Request) *appError {

	var ids []string

	urlID := mux.Vars(r)["fileID"]

	ids = append(ids, urlID)

	body, _ := ioutil.ReadAll(r.Body)

	var idData map[string][]string

	err := json.Unmarshal(body, &idData)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	ids = append(ids, idData["contents"]...)

	basePath := "/var/tmp/dantser-tmp/" + urlID

	err = os.MkdirAll(basePath, 0777)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	for _, i := range ids {
		f.CreateTree(i, urlID)
	}

	status := f.Compress(basePath, basePath+".zip")
	ClearDirectory(basePath)
	os.Remove(basePath)
	if status == true {
		file, err := os.Open(basePath + ".zip")

		defer file.Close()

		if err != nil {

		}

		data, err := ioutil.ReadAll(file)

		if err != nil {

		}
		w.Header().Set("Content-Disposition", "attachment; filename="+urlID+".zip")
		w.Header().Set("Content-Type", "application/zip")
		w.Write(data)
		os.Remove(basePath + ".zip")
	}
	return nil
}

func DownloadFile(w http.ResponseWriter, r *http.Request) *appError {
	fileid := mux.Vars(r)["fileID"]

	if bson.IsObjectIdHex(fileid) {
		data, err := f.FindById(fileid)
		if err != nil {
			return appErrorf(err, "%+v")
		}
		var fm Metadata
		err = data.GetMeta(&fm)

		if err != nil {
			return appErrorf(err, "%+v")
		}

		if fm.Kind == "drive#file" {
			cnt, err := ioutil.ReadAll(data)
			if err != nil {
				return appErrorf(err, "%+v")
			}
			defer data.Close()
			w.Header().Set("Content-Disposition", "attachment; filename="+data.Name())
			w.Header().Set("Content-Type", data.ContentType())
			w.Write(cnt)
			return nil

		} else {

			err := f.CreateTree(fileid, fileid)
			if err != nil {
				return appErrorf(err, "%+v")
			}
			path := "/var/tmp/dantser-tmp/" + fileid

			status := f.Compress(path, path+".zip")

			ClearDirectory(path)

			os.Remove(path)

			if status == true {
				file, err := os.Open(path + ".zip")

				defer file.Close()

				if err != nil {
					return appErrorf(err, "%+v")
				}

				data, err := ioutil.ReadAll(file)

				if err != nil {
					return appErrorf(err, "%+v")
				}

				err = os.Remove(path + ".zip")

				if err != nil {
					return appErrorf(err, "%+v")
				}

				w.Header().Set("Content-Disposition", "attachment; filename="+fm.Name+".zip")
				w.Header().Set("Content-Type", "application/zip")
				w.Write(data)
			}
		}

	}
	return nil
}

func SaveVideo(w http.ResponseWriter, r *http.Request) *appError {
	r.ParseMultipartForm(0)

	metadata := r.Form.Get("metadata")

	file, head, err := r.FormFile("file")
	if err != nil {
		return appErrorf(err, "%#v", err)
	}
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return appErrorf(err, "%#v", err)
	}

	camera, err := c.SaveVideo([]byte(metadata), head, buf)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	if camera != nil {
		w.Header().Set("Content-Type", "application/json")

		res, err := json.Marshal(camera)
		if err != nil {
			return appErrorf(err, "%+v", err)
		}

		w.Write(res)

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil

}

func GetFiles(w http.ResponseWriter, r *http.Request) *appError {
	vars := mux.Vars(r)
	camId := vars["camId"]

	var beginIntTime, endIntTime time.Time
	var checkstart, checkend bool

	begin := r.URL.Query().Get("begin")

	if begin != "" {
		beginInt, err := strconv.ParseInt(begin, 10, 64)
		if err != nil {
			return appErrorf(err, "1 %+v", err)
		}
		checkstart = true
		beginIntTime = time.Unix(beginInt, 0)
	}

	end := r.URL.Query().Get("end")
	if end != "" {
		endInt, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			return appErrorf(err, "2 %+v", err)
		}
		checkend = true
		endIntTime = time.Unix(endInt, 0)
	}

	res, err := c.Get(camId, beginIntTime, endIntTime, checkstart, checkend)
	if err != nil {
		return appErrorf(err, "3 %+v", err)
	}
	r1, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(r1)
	return nil
}

func GetVideo(w http.ResponseWriter, r *http.Request) *appError {
	vars := mux.Vars(r)

	camId := vars["camId"]
	fileName := vars["fileName"]

	file, err := c.FindTSFile(camId, fileName)

	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	err = file.Close()
	if err != nil {
		return appErrorf(err, "%+v", err)
	}

	cnt_type := file.ContentType()

	w.Header().Set("Content-Type", cnt_type)
	w.Write(buf)

	return nil

}

func checkChunkIndex(contentRange string) bool {

	re := regexp.MustCompile(`([a-z]+) ([[:alnum:]]+)-([[:alnum:]]+)/([[:alnum:]]+)`)

	component := re.FindStringSubmatch(contentRange)

	filesize, _ := strconv.Atoi(component[4])

	last_chunk, _ := strconv.Atoi(component[3])

	return !(filesize-last_chunk-1 > 0)

}

func CreateStorage(w http.ResponseWriter, r *http.Request) *appError {
	var sd StorageData

	beingID := r.Header["X-Being-Id"][0]

	size := r.URL.Query().Get("size")

	sd.Owner, _ = strconv.Atoi(beingID)

	sizeInt, err := strconv.Atoi(size)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	sd.Size = sizeInt

	id, err := sd.Create(sd)

	if err != nil {
		w.WriteHeader(400)
		return appErrorf(err, "%+v")
	}

	w.WriteHeader(201)
	w.Write([]byte(`{id:` + id + `}`))
	return nil

	return nil
}

func UpdateStorage(w http.ResponseWriter, r *http.Request) *appError {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return appErrorf(err, "%+v")
	}

	var data map[string]string

	json.Unmarshal(body, &data)

	var sd StorageData

	sd.Size, err = strconv.Atoi(data["size"])

	active, err := sd.GetActive(data["beingID"])
	sd.Update(sd, active.Id)

	return nil
}

func DeleteStorage(w http.ResponseWriter, r *http.Request) *appError {
	id := mux.Vars(r)["storageID"]
	beingId, err := strconv.Atoi(r.Header["X-being-Id"][0])

	if err != nil {
		return appErrorf(err, "%+v")
	}

	count, err := Connection.Session.DB("cctv_storage").C("storages").Find(bson.M{"_id": id, "owner": beingId}).Count()

	if count == 0 || err != nil {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	err = s.Delete(id)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	w.WriteHeader(200)
	return nil
}

func ClearStorage(w http.ResponseWriter, r *http.Request) *appError {

	id := mux.Vars(r)["storageID"]

	beingId, err := strconv.Atoi(r.Header["X-Being-Id"][0])

	if err != nil {
		return appErrorf(err, "%+v")
	}

	count, err := Connection.Session.DB("cctv_storage").C("storages").Find(bson.M{"_id": id, "owner": beingId}).Count()

	if count == 0 || err != nil {
		return appErrorf(mgo.ErrNotFound, "%+v")
	}

	err = s.Clear(id)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	w.WriteHeader(200)
	return nil
}

func ActivateStorage(w http.ResponseWriter, r *http.Request) *appError {
	id := mux.Vars(r)["storageID"]

	beingId := r.Header["X-Being-Id"][0]

	err := s.Activate(id, beingId)

	if err != nil {
		return appErrorf(err, "%+v")
	}

	w.WriteHeader(200)

	return nil
}
