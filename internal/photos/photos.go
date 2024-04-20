package photos

import (
	"database/sql"
	"image"
	"io/fs"
	"net/http"
)

var StaticPhotosFS = http.FileServer(http.Dir("~/projects/scoretable/back_end/static"))

type Photo struct {
	img    image.Image
	userID int64
}

func (p *Photo) Store(store PhotoStore) (string, error) {
	url, err := store.Store(p)
	if err != nil {
		return "", err
	}
	return url, nil
}

type PhotoStore interface {
	Store(img *Photo) (string, error)
}

type StaticPhotoStore struct {
	fs fs.FS
	db *sql.DB
}

func (s *StaticPhotoStore) Store(img *Photo) (string, error) {
	return "", nil
}
