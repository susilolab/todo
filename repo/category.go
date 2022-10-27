package repo

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"todo/models"
	"todo/utils"

	"github.com/boltdb/bolt"
)

type CategoryRepo struct {
	Db *bolt.DB
}

func NewCategoryRepo(db *bolt.DB) *CategoryRepo {
	return &CategoryRepo{Db: db}
}

func (cr *CategoryRepo) FindAll() ([]models.Category, error) {
	db := cr.Db
	rows := make([]models.Category, 0)
	err := db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("category"))
        if err != nil {
            return err
        }
		b.ForEach(func(k, v []byte) error {
			category := models.Category{}
			_ = json.Unmarshal(v, &category)
			rows = append(rows, category)
			return nil
		})
		return nil
	})

	return rows, err
}

func (cr *CategoryRepo) FindOne(id int) (*models.Category, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return nil, err
	}

	category := &models.Category{}
	db := cr.Db
	err = db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("category"))
        if err != nil {
            return err
        }
		v := b.Get(idb)

		if v == nil {
			return fmt.Errorf("Data tidak ditemukan.")
		}

		err = json.Unmarshal(v, category)
		if err != nil {
			return err
		}
		return nil
	})

	return category, err
}

func (cr *CategoryRepo) Create(category *models.Category) (*models.Category, error) {
	db := cr.Db
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("category"))
        if err != nil {
            return err
        }

		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		category.ID = int64(id)
		buf, err := json.Marshal(category)
		if err != nil {
			return err
		}

		bId := new(bytes.Buffer)
		binary.Write(bId, binary.LittleEndian, id)

		return b.Put(bId.Bytes(), buf)
	})
	return category, err
}

func (cr *CategoryRepo) Update(id int, category *models.Category) (*models.Category, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return category, err
	}

	err = cr.Db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("category"))
        if err != nil {
            return err
        }
		c := b.Cursor()

		k, v := c.Seek(idb)
		if k == nil {
			return fmt.Errorf("id %d tidak ditemukan.", id)
		}
		err = json.Unmarshal(v, category)

		return nil
	})
	return category, nil
}

func (cr *CategoryRepo) Delete(id int) error {
	db := cr.Db
	bId, err := utils.IntToBytes(int64(id))
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("category"))
        if err != nil {
            return err
        }
		c := b.Cursor()

		k, _ := c.Seek(bId)
		if k != nil {
			return b.Delete(k)
		} else {
			return fmt.Errorf("Data tidak ditemukan, hapus dibatalkan.")
		}
	})
	return err
}
