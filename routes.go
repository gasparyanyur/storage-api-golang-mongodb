package main

import (
	"net/http"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	routers := mux.NewRouter().StrictSlash(true)
	routers.Methods("GET").Path("/").Handler(appHandler(Index))

	// FILES
	f := routers.PathPrefix("/drive/" + ProtocolVersion).Subrouter()
	f.Methods("GET").Path("/files").Handler(appHandler(FileList))
	f.Methods("GET").Path("/files/deleted").Handler(appHandler(GetDeleted))
	f.Methods("GET").Path("/file/{fileID}/tree").Handler(appHandler(GetTree))
	f.Methods("GET").Path("/files/{fileID}/restore").Handler(appHandler(Restore))
	f.Methods("GET").Path("/files/{fileID}").Queries("alt", "textfile").Handler(appHandler(TextFileInfo))

	//DOWNLOAD
	f.Methods("GET").Path("/files/{fileID}").Queries("alt", "media").Handler(appHandler(DownloadFile))
	f.Methods("POST").Path("/files/{fileID}").Queries("alt", "media").Handler(appHandler(BulkDownloadFile))
	f.Methods("GET").Path("/file/{fileId}").Handler(appHandler(FileGet))
	//DELETE
	f.Methods("DELETE").Path("/files/{Id}").Queries("type", "force").Handler(appHandler(ForceDelete))
	f.Methods("DELETE").Path("/files/{Id}").Handler(appHandler(Delete))

	//Storages
	s := routers.PathPrefix("/storage").Subrouter()
	s.Methods("GET").Path("").Handler(appHandler(CreateStorage))
	s.Methods("POST").Path("").Handler(appHandler(UpdateStorage))
	s.Methods("DELETE").Path("/{storageID}").Handler(appHandler(ClearStorage))
	s.Methods("DELETE").Path("/{storageID}/delete").Handler(appHandler(DeleteStorage))
	s.Methods("POST").Path("/{storageID}/activate").Handler(appHandler(ActivateStorage))

	// UPLOAD
	u := routers.PathPrefix("/upload/drive/" + ProtocolVersion).Subrouter()
	u.Methods("POST").Path("/files/{folderID}").Queries("uploadType", "textfile").Handler(appHandler(UploadTextFile))
	u.Methods("PUT").Path("/files/{fileID}").Queries("uploadType", "textfile").Handler(appHandler(UpdateTextFile))
	//uploading a file
	u.Methods("POST").Path("/files").Queries("uploadType", "media").Handler(appHandler(SimpleUpload))
	u.Methods("POST").Path("/files").Queries("uploadType", "multipart").Handler(appHandler(MultiPartUpload))
	u.Methods("POST").Path("/files").Queries("uploadType", "resumable").Handler(appHandler(ResumableUploadFile))
	u.Methods("PUT").Path("/files/{fileID}").Queries("uploadType", "resumable").Handler(appHandler(ResumableUploadFile))
	//copying a file
	u.Methods("POST").Path("/files/{fileID}/copy").Handler(appHandler(CopyFile))
	// update file metadata
	u.Methods("PUT").Path("/files/{fileID}").Handler(appHandler(UpdateFile))

	// CAMERAS
	c := routers.PathPrefix("/cctv/" + ProtocolVersion).Subrouter()
	c.Methods("POST").Path("/cameras/video").Handler(appHandler(SaveVideo))
	c.Methods("GET").Path("/cameras/list/{camId}").Handler(appHandler(GetFiles))

	// CAMERAS RESPONSE
	cr := routers.PathPrefix("/cctv/").Subrouter()
	cr.Methods("GET").Path("/video/{camId}/{fileName}").Handler(appHandler(GetVideo))

	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, routers))
	return routers
}
