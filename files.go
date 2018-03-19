package main

import (
	"net/http"
	"time"

	"errors"

	"io/ioutil"
	"strings"

	"os"

	"archive/zip"
	"io"
	"path/filepath"

	"strconv"

	"log"

	"gopkg.in/h2non/filetype.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MimeExt struct {
	Mime string
	Ext  string
}

type Datas map[string]interface{}

var dirs []string

var mgoFile *mgo.GridFile

type File struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	ChunkSize   uint32        `json:"chunkSize" bson:"chunkSize"`
	ContentType string        `json:"contentType" bson:"contentType"`
	Filename    string        `json:"filename" bson:"filename"`
	Length      uint32        `json:"length" bson:"length"`
	Md5         string        `json:"md5" bson:"md5"`
	Metadata    Metadata      `json:"metadata" bson:"metadata"`
	UploadDate  time.Time     `json:"uploadDate" bson:"uploadDate"`
}

type TextFile struct {
	Id       bson.ObjectId   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string          `json:"name,omitempty" bson:"name,omitempty"`
	OwnerID  string          `json:"ownerid,omitempty" bson:"ownerid,omitempty"`
	Content  []byte          `json:"content,omitempty" bson:"content,omitempty"`
	Text     string          `json:"text,omitempty" bson:"-"`
	Parents  []bson.ObjectId `json:"parentid,omitempty" bson:"parentid,omitempty"`
	Created  time.Time       `json:"created,omitempty" bson:"created,omitempty"`
	Modified time.Time       `json:"modified,omitempty" bson:"modified,omitempty"`
}

type Metadata struct {
	Name      string          `json:"name,omitempty" bson:"-"`
	Trashed   bool            `json:"trashed,omitempty" bson:"trashed,omitempty"`
	OwnerId   int             `json:"oid,omitempty" bson:"oid,omitempty"`
	StorageId bson.ObjectId   `json:"sid,omitempty" bson:"sid"`
	Kind      string          `json:"kind,omitempty" bson:"kind,omitempty"`
	Parent    string          `json:"parent,omitempty" bson:"-"`
	Parents   []bson.ObjectId `json:"parents,omitempty" bson:"parents"`
}

type ChildParentsData struct {
	Id      bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	Trashed bool          `json:"trashed" bson:"trashed"`
}

type TreeData struct {
	Fid      bson.ObjectId      `json:"fid,omitempty" bson:"fid,omitempty"`
	Parents  []ChildParentsData `json:"pr" bson:"pr"`
	Children []ChildParentsData `json:"ch" bson:"ch"`
}

func (f *File) List() ([]File, error) {
	iter := Connection.Gfs.Find(nil).Iter()
	var file File
	var res []File
	for iter.Next(&file) {
		res = append(res, file)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return res, nil
}

func (f *File) Childs(id bson.ObjectId) ([]File, error) {
	var tree []ChildParentsData
	var ids []bson.ObjectId

	pipeline := []bson.M{
		{"$match": bson.M{"fid": id}},
		{"$unwind": "$ch"},
		{"$match": bson.M{"ch.trashed": false}},
		{"$replaceRoot": bson.M{"newRoot": "$ch"}}}

	err := Connection.Session.DB("cctv_storage").C("tree").Pipe(pipeline).All(&tree)

	if err != nil {
		return nil, err
	}

	for _, v := range tree {
		ids = append(ids, v.Id)
	}

	var fs []File

	err = Connection.Gfs.Find(bson.M{"_id": bson.M{"$in": ids}}).All(&fs)

	if err != nil {
		return nil, err
	}

	return fs, nil

}

func (file *File) ForceDelete(id bson.ObjectId) error {

	var tData struct {
		Id       bson.ObjectId `bson:"_id"`
		Children []*File       `bson:"p"`
	}

	pipeline := []bson.M{
		{"$unwind": bson.M{
			"path":                       "$metadata.parents",
			"preserveNullAndEmptyArrays": true,
		}},
		{"$graphLookup": bson.M{
			"from":             "fs.files",
			"startWith":        "$_id",
			"connectFromField": "_id",
			"connectToField":   "metadata.parents",
			"as":               "p",
		}},
		{"$match": bson.M{"_id": id}},
	}

	Connection.Session.DB("cctv_storage").C("fs.files").Pipe(pipeline).One(&tData)

	for _, k := range tData.Children {
		err := Connection.Gfs.RemoveId(k.Id)
		if err != nil {
			return err
		}
	}

	err := Connection.Gfs.RemoveId(id)

	if err != nil {
		return err
	}

	return nil

}

func (file *File) Restore(id bson.ObjectId) error {

	err := Connection.Session.DB("cctv_storage").C("fs.files").UpdateId(
		id, bson.M{"$set": bson.M{"metadata.trashed": false}})

	if err != nil {
		return err
	}
	return nil
}

func (f *File) Get(fileId string) (*File, error) {
	var file File
	err := Connection.Gfs.Find(bson.M{"_id": bson.ObjectIdHex(fileId)}).One(&file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *File) CreateTextFile(dt TextFile) string {
	id := bson.NewObjectId()
	dt.Id = id
	err := Connection.Session.DB("cctv_storage").C("text_files").Insert(dt)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return id.Hex()
}

func (f *File) UpdateTexFile(id bson.ObjectId, dt TextFile) bool {
	err := Connection.Session.DB("cctv_storage").C("text_files").UpdateId(id, bson.M{"$set": dt})
	if err != nil {
		return false
	}
	return true
}

func (f *File) SimpleCreate(buf []byte, beingID string) (*File, error) {
	mimeExt, err := getMimeExt(buf[:512])
	if err != nil {
		return nil, errors.New("unknown file type")
	}
	id := bson.NewObjectId()

	filename := "Simple_" + id.Hex()
	if mimeExt.Ext != "" {
		filename += "." + mimeExt.Ext
	}

	mfile, err := Connection.Gfs.Create(filename)

	var meta Metadata

	var s StorageData

	being, err := strconv.Atoi(beingID)

	if err != nil {
		return nil, err
	}

	meta.OwnerId = being

	active, err := s.GetActive(beingID)

	if err != nil {
		return nil, err
	}

	meta.StorageId = active.Id

	if err != nil {
		return nil, err
	}
	mfile.SetId(id)

	mfile.SetMeta(meta)

	if mimeExt.Mime != "" {
		mfile.SetContentType(mimeExt.Mime)
	}

	_, err = mfile.Write(buf)
	if err != nil {
		return nil, err
	}

	err = mfile.Close()
	if err != nil {
		return nil, err
	}

	err = AddToTree(id, id)
	if err != nil {
		return nil, err
	}

	s.Used = len(buf)

	s.Update(s, active.Id)

	return &File{
		Id:          mfile.Id().(bson.ObjectId),
		Filename:    mfile.Name(),
		ContentType: mfile.ContentType(),
	}, nil
}

func (f *File) MultiPartCreate(buf []byte, meta Metadata, beingID string) (*File, error) {
	mimeExt, err := getMimeExt(buf[:512])
	if err != nil {
		return nil, errors.New("unknown file type")
	}

	filename := meta.Name

	id := bson.NewObjectId()

	var pid bson.ObjectId

	mfile, err := Connection.Gfs.Create(filename)

	if err != nil {
		return nil, err
	}

	var sd StorageData

	active, err := sd.GetActive(beingID)

	mfile.SetId(id)

	meta.Kind = "drive#file"
	meta.StorageId = active.Id

	mfile.SetMeta(meta)

	if mimeExt.Mime != "" {
		mfile.SetContentType(mimeExt.Mime)
	}

	if meta.Parent != "" {
		pid = bson.ObjectIdHex(meta.Parent)
		meta.Parents = append(meta.Parents, pid)
	}

	_, err = mfile.Write(buf)

	if err != nil {
		return nil, err
	}

	err = mfile.Close()
	if err != nil {
		return nil, err
	}

	sd.Size = len(buf)

	sd.Update(sd, active.Id)

	return &File{
		Id:          mfile.Id().(bson.ObjectId),
		Filename:    mfile.Name(),
		ContentType: mfile.ContentType(),
	}, nil
}

func (f *File) ResumableServeMetadata(metadata Metadata, buf []byte) (string, error) {

	filename := metadata.Name

	var err error

	var pid bson.ObjectId

	mgoFile, err = Connection.Gfs.Create(filename)

	if err != nil {
		return "", err
	}

	id := bson.NewObjectId()

	mgoFile.SetId(id)

	if metadata.Parent != "" {
		pid = bson.ObjectIdHex(metadata.Parent)
		metadata.Parents = append(metadata.Parents, pid)
	}

	metadata.Kind = "drive#file"
	mgoFile.SetMeta(metadata)

	return id.Hex(), nil

}

func (f *File) ResumableServeChunkedFile(buf []byte, contentType string, length int, close bool) (int, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	mgoFile.SetContentType(contentType)

	_, err := mgoFile.Write(buf)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	if close == true {
		err = mgoFile.Close()

		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusCreated, nil
	}

	return http.StatusPermanentRedirect, err
}

func (f *File) CreateFolder(data Metadata, beingId string) (*File, error) {
	var pid bson.ObjectId

	file, err := Connection.Gfs.Create(data.Name)

	if err != nil {
		return nil, err
	}

	id := bson.NewObjectId()

	file.SetId(id)

	if data.Parent != "" {
		pid = bson.ObjectIdHex(data.Parent)

		count, err := Connection.Gfs.Find(bson.M{"_id": pid, "metadata.kind": "drive#folder"}).Count()

		if count == 0 || err != nil {
			return nil, mgo.ErrNotFound
		}

		data.Parents = append(data.Parents, pid)

	}

	data.Kind = "drive#folder"

	active, err := s.GetActive(beingId)

	if err != nil {
		return nil, err
	}

	data.StorageId = active.Id

	file.SetMeta(data)

	file.SetContentType("application/vnd.dantser-apps.folder")

	file.Close()

	return &File{
		Id:          file.Id().(bson.ObjectId),
		Filename:    file.Name(),
		ContentType: file.ContentType(),
	}, nil
}

/*
  TODO : Improve getting folder tree
*/
func (f *File) CreateTree(id string, rootPath string) error {
	var documentRoot = "/var/tmp/dantser-tmp" + "/" + rootPath

	parentId := bson.ObjectIdHex(id)

	var fData File

	err := Connection.Session.DB("cctv_storage").C("fs.files").Find(bson.M{"_id": parentId}).One(&fData)

	if err != nil {
		return err
	}

	dirExists, _ := checkFileOrDirectory(documentRoot)

	if dirExists != true {
		err := os.MkdirAll(documentRoot, 0777)

		if err != nil {
			return err
		}
	}

	if fData.Metadata.Kind == "drive#folder" {

		created := createDirectory(fData, nil, documentRoot)

		if created == false {
			return errors.New("Can not create directory")
		}

		documentRoot = documentRoot + "/" + fData.Filename

		var data = struct {
			Id    bson.ObjectId `json:"_id" bson:"_id"`
			Name  bson.ObjectId `json:"filename" bson:"filename"`
			Child []File        `json:"child" bson:"child"`
		}{}

		pipeline := []bson.M{
			bson.M{"$match": bson.M{"$or": []bson.M{bson.M{"metadata.trashed": bson.M{"$exists": false}}, bson.M{"metadata.trashed": false}}}},
			bson.M{"$unwind": bson.M{
				"path":                       "$metadata.parents",
				"preserveNullAndEmptyArrays": true,
			}},
			bson.M{"$graphLookup": bson.M{
				"from":             "fs.files",
				"startWith":        "$_id",
				"connectFromField": "_id",
				"connectToField":   "metadata.parents",
				"as":               "child",
			}},
			bson.M{"$match": bson.M{"_id": parentId}},
		}

		Connection.Session.DB("cctv_storage").C("fs.files").Pipe(pipeline).One(&data)

		children := data.Child

		for _, f := range children {
			if checkParent(parentId, f) {
				if f.Metadata.Kind == "drive#folder" {
					createDirectory(f, children, documentRoot)
				} else {
					createFile(f, documentRoot)
				}
			}
		}

	} else {
		err = createFile(fData, documentRoot)

		if err != nil {
			return err
		}
	}
	return nil

}

func (f *File) Update(id bson.ObjectId, data map[string]interface{}) error {

	err := Connection.Session.DB("cctv_storage").C("fs.files").UpdateId(id, data)

	if err != nil {
		return err
	}

	return nil
}

func (f *File) GetDeleted(being int) (data []File, err error) {

	rootPipeline := []bson.M{
		{"$match": bson.M{"metadata.oid": being}},
		{"$lookup": bson.M{
			"from":         "fs.files",
			"localField":   "metadata.parents",
			"foreignField": "_id",
			"as":           "p",
		}},
		{"$match": bson.M{
			"$or": []bson.M{
				{"metadata.trashed": true, "p": bson.M{"$size": 0}},
				{"metadata.trashed": true, "$or": []bson.M{
					{"p.metadata.trashed": bson.M{"$exists": false}},
					bson.M{"p.metadata.trashed:": false}},
				},
			},
		}},
	}

	err = Connection.Session.DB("cctv_storage").C("fs.files").Pipe(rootPipeline).All(&data)

	return
}

func (f *File) GetTree(id bson.ObjectId, parentId bson.ObjectId) ([]bson.ObjectId, error) {
	var dt []struct {
		Id      bson.ObjectId              `bson:"_id"`
		Pid     bson.ObjectId              `bson:"pid"`
		Fid     bson.ObjectId              `bson:"fid"`
		Parents []map[string]bson.ObjectId `bson:"parents"`
	}

	query := []bson.M{{
		"$graphLookup": bson.M{ // lookup the documents table here
			"from":             "tree",
			"startWith":        "$fid",
			"connectFromField": "fid",
			"connectToField":   "pid",
			"as":               "parents",
		}},
	}

	err := Connection.Session.DB("cctv_storage").C("tree").Pipe(query).All(&dt)

	var ids []bson.ObjectId

	if err != nil {
		return ids, err
	}

	var parents []map[string]bson.ObjectId

	for _, v := range dt {
		if v.Fid == id && v.Pid == parentId {
			parents = v.Parents
			break
		}

	}

	for _, v := range parents {
		ids = append(ids, v["_id"])
	}
	return ids, nil
}

func (f *File) ForceDeleteFolder(id bson.ObjectId, parentId bson.ObjectId) error {

	ids, err := f.GetTree(id, parentId)

	if err != nil {
		return err
	}
	_, err = Connection.Session.DB("cctv_storage").C("tree").RemoveAll(bson.M{"_id": bson.M{"$in": ids}})

	if err == nil {
		for _, v := range ids {
			Connection.Gfs.RemoveId(v)
		}
	}

	if err != nil {
		return err
	}

	return nil

}

func (f *File) Delete(id bson.ObjectId) error {
	err := Connection.Session.DB("cctv_storage").C("fs.files").UpdateId(
		id, bson.M{"$set": bson.M{"metadata.trashed": true}})

	if err != nil {
		return err
	}
	return nil

}

func (f *File) Copy(id bson.ObjectId) error {

	var fm Metadata

	mfile, err := Connection.Gfs.OpenId(id)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(mfile)

	if err != nil {
		return err
	}

	err = mfile.GetMeta(&fm)

	if err != nil {
		return err
	}

	contents := strings.Split(mfile.Name(), ".")

	var filename string

	if len(contents) == 1 {
		filename = contents[0] + " (Copy)"
	} else {
		contents[len(contents)-2] = contents[len(contents)-2] + " (Copy) ."
		filename = strings.Join(contents, "")
	}

	copy, err := Connection.Gfs.Create(filename)

	if err != nil {
		return err
	}

	_, err = copy.Write(data)

	if err != nil {
		return err
	}

	copy.SetMeta(fm)

	err = copy.Close()

	if err != nil {
		return err
	}

	return nil
}

func (f *File) FindById(id string) (*mgo.GridFile, error) {
	fileid := bson.ObjectIdHex(id)

	data, err := Connection.Gfs.OpenId(fileid)

	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *File) Compress(source, target string) bool {
	zipfile, err := os.Create(target)

	if err != nil {
		return false
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)

	defer archive.Close()

	info, err := os.Stat(source)

	if err != nil {
		return false
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return true
}

func checkFileOrDirectory(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkDupliacte(path string) int {
	var count int = 0
	for _, k := range dirs {
		if k == path {
			count++
		}
	}
	return count
}

func dirNameRender(name string, count int) string {

	components := strings.Split(name, ".")

	if len(components) > 1 {
		components[len(components)-2] = components[len(components)-2] + " (" + strconv.Itoa(count+2) + ")"
	} else {
		components[0] = components[0] + " (" + strconv.Itoa(count+2) + ")"
	}
	name = strings.Join(components, ".")
	return name
}

func ClearDirectory(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func getMimeExt(buf []byte) (*MimeExt, error) {
	kind, err := filetype.Match(buf)
	if err != nil {
		return nil, err
	}
	if kind.MIME.Value != "" {
		return &MimeExt{
			kind.MIME.Value,
			kind.Extension,
		}, nil
	}

	if subType := http.DetectContentType(buf); subType != "" {
		return &MimeExt{
			subType,
			"",
		}, nil
	}

	return &MimeExt{}, nil
}

func AddToTree(fid, pid bson.ObjectId) error {
	var tData TreeData

	var pData ChildParentsData

	tData.Fid = fid

	if fid != pid {
		pData.Id = pid
		tData.Parents = append(tData.Parents, pData)
	}

	err := Connection.Session.DB("cctv_storage").C("tree").Insert(tData)

	if err != nil {
		return err
	}

	if fid != pid {
		var tChild ChildParentsData

		tChild.Id = fid

		err = Connection.Session.DB("cctv_storage").C("tree").Update(bson.M{"fid": pid}, bson.M{"$push": bson.M{"ch": tChild}})

		if err != nil {
			return err
		}
	}

	return nil
}

func checkParent(parentId bson.ObjectId, child File) bool {

	for _, id := range child.Metadata.Parents {
		if id == parentId {
			return true
		}
	}
	return false
}

func createDirectory(parent File, children []File, path string) (created bool) {
	allPath := path + "/" + parent.Filename

	dirExists, _ := checkFileOrDirectory(allPath)

	if dirExists == true {
		count := checkDupliacte(allPath)
		allPath = allPath + "(" + strconv.Itoa(count) + ")"
	}

	err := os.MkdirAll(allPath, 0777)

	if err != nil {
		return false
	}

	dirs = append(dirs, allPath)

	for _, c := range children {
		if checkParent(parent.Id, c) {
			if c.Metadata.Kind == "drive#folder" {
				createDirectory(c, children, allPath)
			} else {
				createFile(c, allPath)
			}
		}
	}

	return true
}

func createFile(file File, path string) error {

	allPath := path + "/" + file.Filename

	count := checkDupliacte(allPath)

	if count > 0 {
		allPath = dirNameRender(allPath, count)
	}

	osFile, err := os.Create(allPath)

	if err != nil {
		return err
	}

	mFile, err := Connection.Gfs.OpenId(file.Id)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(mFile)

	if err != nil {
		return err
	}

	_, err = osFile.Write(data)

	if err != nil {
		return err
	}

	err = osFile.Close()

	if err != nil {
		return err
	}

	return nil
}
