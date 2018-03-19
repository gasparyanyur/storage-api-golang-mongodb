package main

import (
	"testing"

	"log"

	"io/ioutil"

	"math"
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//-dbhost 10.99.0.198 -dbname cctv_storage -host :8019

var cred *mgo.Credential

var id, txtId, fid, stId, rootid bson.ObjectId
var basePath string

func TestStorageData_Create(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	var sd StorageData
	sd.Size = 2000000000
	sd.Owner = 123
	str, err := s.Create(sd)

	if err != nil {
		t.Fatal(err)
	}
	if bson.IsObjectIdHex(str) {
		t.Log("TestStorageData_Create : Success ")
	} else {
		t.Log("TestStorageData_Create : Wrong response")
	}

}

func TestStorageData_CreateSecond(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	var sd StorageData
	sd.Size = 2000000000
	sd.Owner = 123
	str, err := s.Create(sd)

	if err != nil {
		t.Fatal(err)
	}
	if bson.IsObjectIdHex(str) {
		stId = bson.ObjectIdHex(str)
		t.Log("TestStorageData_Create : Success ")
	} else {
		t.Log("TestStorageData_Create : Wrong response")
	}

}

func TestActivateStorage(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	err := s.Activate(stId.Hex(), "123")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestActivateStorage : Success")
}

func TestFile_CreateFolder(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	var mt Metadata

	mt.Name = "Folder 1"
	mt.StorageId = stId
	mt.OwnerId = 123

	file, err := f.CreateFolder(mt, "123")

	if err != nil {
		t.Fatal(err)
	}

	if file.Filename != "" && file.ContentType != "" {
		rootid = file.Id
		t.Log("TestFile_CreateFolder: Success")
	} else {
		t.Log("TestFile_CreateFolder: Wrong Response")
	}
}

func TestFile_CreateFolderWithParent(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	var mt Metadata

	mt.Name = "SubFolder 1"

	mt.Parent = rootid.Hex()

	file, err := f.CreateFolder(mt, "123")

	if err != nil {
		t.Fatal(err)
	}

	if file.Filename != "" && file.ContentType != "" {
		fid = file.Id
		t.Log("TestFile_CreateFolder: Success")
	} else {
		t.Log("TestFile_CreateFolder: Wrong Response")
	}
}

func TestSimpleUpload(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	osFile, err := ioutil.ReadFile("./demo/test.jpg")

	if err != nil {
		t.Fatal(err)
	}

	file, err := f.SimpleCreate(osFile, "123")

	if err != nil {
		t.Fatal(err)
	}

	if file.Filename != "" && file.ContentType != "" {
		id = file.Id
		t.Log("TestSimpleUpload: Success")
	} else {
		t.Log("TestSimpleUpload: Wrong Response")
	}

}

func TestFile_MultiPartCreate(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	var mt Metadata

	osFile, err := ioutil.ReadFile("./demo/test.jpg")

	mt.Name = "file001.jpg"
	mt.Parent = fid.Hex()

	file, err := f.MultiPartCreate(osFile, mt, "123")

	if err != nil {
		t.Fatal(err)
	}

	if file.Filename != "" && file.ContentType != "" {
		id = file.Id
		t.Log("TestFile_MultiPartCreate : Success")
	} else {
		t.Log("TestFile_MultiPartCreate : Wrong response")
	}
}

func TestFile_ResumableServeMetadata(t *testing.T) {

	Connection, _ = newMongoDB(config.dbHost, cred)

	var fm Metadata

	fm.Name = "file.png"

	fm.StorageId = bson.NewObjectId()

	id, err := f.ResumableServeMetadata(fm, nil)

	if err != nil {
		t.Fatal(err)
	}

	if bson.IsObjectIdHex(id) == true {
		t.Log("TestFile_ResumableServeMetadata : Success")
	} else {
		t.Log("TestFile_ResumableServeMetadata : Wrong response")
	}

	osFile, err := os.Open("./demo/test.jpg")

	if err != nil {
		log.Fatal(err)
	}

	fileinfo, err := osFile.Stat()

	if err != nil {
		log.Fatal(err)
	}

	const fileChunk = 0.25 * (1 << 20)

	filesize := fileinfo.Size()

	totalPartsNum := uint64(math.Ceil(float64(filesize) / float64(fileChunk)))

	var sent int

	var close bool

	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(filesize-int64(i*fileChunk))))

		sent = sent + partSize
		partBuffer := make([]byte, partSize)

		if int(sent) == int(filesize) {
			close = true
		}

		osFile.Read(partBuffer)

		_, err := f.ResumableServeChunkedFile(partBuffer, "image/jpg", partSize, close)

		if err != nil {
			t.Fatal(err)
		}

	}

}

func TestFile_Copy(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	err := f.Copy(id)

	if err != nil {
		t.Fatal(err)
	}
	t.Log("TestFile_Copy : Success")
}

func TestFile_CreateTextFile(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	var dt TextFile

	dt.Name = "asdasdasd.txt"
	dt.OwnerID = "123"
	dt.Parents = append(dt.Parents, fid)

	str := f.CreateTextFile(dt)

	if bson.IsObjectIdHex(str) {
		txtId = bson.ObjectIdHex(str)
		t.Log("TestFile_CreateTextFile : Success")
	} else {
		t.Log("TestFile_CreateTextFile : Wrong Response")
	}
}

func TestFile_UpdateTexFile(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	file, err := ioutil.ReadFile("./demo/test.txt")

	if err != nil {
		log.Fatal(err)
	}

	var td TextFile

	td.Content = file

	upd := f.UpdateTexFile(txtId, td)

	if upd == true {
		t.Log("TestFile_UpdateTexFile : Success")
	} else {
		t.Log("TestFile_UpdateTexFile : Fail")
	}
}

func TestFile_CreateTree(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	treeId := fid.Hex()

	err := f.CreateTree(treeId, treeId)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestFile_CreateTree : Success")

}

func TestFile_Compress(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	compressLink := "/var/tmp/dantser-tmp/folder.pdf"
	compressed := f.Compress(basePath, compressLink)

	file, err := os.Open(compressLink)

	if err != nil {
		t.Fatal(err)
	}

	fileinfo, err := file.Stat()

	if err != nil {
		t.Fatal(err)
	}

	if fileinfo.Size() == 0 {
		t.Fatal("Wrong file")
	}

	if compressed == true {
		t.Log("TestFile_Compress : Success")
	} else {
		t.Log("TestFile_Compress : Fail")
	}
}

func TestFile_Delete(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	err := f.Delete(rootid)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestFile_Delete : Success")
}

func TestFile_GetDeleted(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)
	beingId := 123

	data, err := f.GetDeleted(beingId)

	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 1 {
		t.Fatal("TestFile_GetDeleted : Wrong Response")
	}

	t.Log("TestFile_GetDeleted : Success ")
}

func TestFile_Restore(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	err := f.Restore(rootid)

	if err != nil {
		t.Fatal(err)
	}

	data, err := f.GetDeleted(123)

	if err != nil {
		t.Fatal(err)
	}

	if len(data) > 0 {
		t.Fatal("TestFile_Restore : Wrong Response")
	}

	t.Log("TestFile_Restore : Success")
}

func TestFile_ForceDelete(t *testing.T) {
	Connection, _ = newMongoDB(config.dbHost, cred)

	err := f.Delete(rootid)

	if err != nil {
		t.Fatal(err)
	}

	err = f.ForceDelete(rootid)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestFile_ForceDelete : Success ")

}
