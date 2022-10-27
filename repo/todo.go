package repo

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
    "errors"

	"todo/models"
	"todo/utils"

	"github.com/boltdb/bolt"
)

type TodoRepo struct {
	Db *bolt.DB
}

func NewTodoRepo(db *bolt.DB) *TodoRepo {
	return &TodoRepo{Db: db}
}

func (tr *TodoRepo) FindAll() ([]models.Todo, error) {
	db := tr.Db
	rows := make([]models.Todo, 0)
	err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("todo"))
        if b == nil {
            return errors.New("Bucket todo tidak ada")
        }

		b.ForEach(func(k, v []byte) error {
			todo := models.Todo{}
            if k != nil {
                _ = json.Unmarshal(v, &todo)
                rows = append(rows, todo)
            }
			return nil
		})
		return nil
	})

	return rows, err
}

func (tr *TodoRepo) FindOne(id int) (*models.Todo, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return nil, err
	}

	todo := &models.Todo{}
	db := tr.Db
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("todo"))
        if b == nil {
            return err
        }
		v := b.Get(idb)

		if v == nil {
			return fmt.Errorf("Data tidak ditemukan.")
		}

		err = json.Unmarshal(v, todo)
		if err != nil {
			return err
		}
		return nil
	})

	return todo, err
}

func (tr *TodoRepo) Create(todo *models.Todo) (*models.Todo, error) {
	db := tr.Db
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("todo"))
        if err != nil {
            return err
        }

		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		todo.ID = int64(id)
		buf, err := json.Marshal(todo)
		if err != nil {
			return err
		}

		bId := new(bytes.Buffer)
		binary.Write(bId, binary.LittleEndian, id)

		return b.Put(bId.Bytes(), buf)
	})
	return todo, err
}

func (tr *TodoRepo) Update(id int, todo *models.Todo) (*models.Todo, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return todo, err
	}

	err = tr.Db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("todo"))
        if err != nil {
            return err
        }
		c := b.Cursor()

		k, v := c.Seek(idb)
		if k == nil {
			return fmt.Errorf("id %d tidak ditemukan.", id)
		}
		err = json.Unmarshal(v, todo)

		return nil
	})
	return todo, nil
}

func (tr *TodoRepo) Delete(id int) error {
	db := tr.Db
	bId, err := utils.IntToBytes(int64(id))
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("todo"))
        if err != nil {
            return err
        }
		c := b.Cursor()

		k, _ := c.Seek(bId)
		if k != nil {
			return b.Delete(k)
		}
		return fmt.Errorf("Data tidak ditemukan, hapus dibatalkan.")
	})
	return err
}

func (tr *TodoRepo) IsTodoDone(id int) (bool, error) {
	result := false
	bId, err := utils.IntToBytes(int64(id))
	if err != nil {
		return false, err
	}

	err = tr.Db.View(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("todo"))
        if err != nil {
            return err
        }
		v := b.Get(bId)

		if v == nil {
			return fmt.Errorf("Data tidak ditemukan")
		}

		todo := models.Todo{}
		json.Unmarshal(v, &todo)
		if todo.Done == 1 {
			result = true
		}
		return nil
	})

	if err != nil {
		return false, err
	}

	return result, nil
}
