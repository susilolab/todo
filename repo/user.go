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

type UserRepo struct {
	Db *bolt.DB
}

func NewUserRepo(db *bolt.DB) *UserRepo {
	return &UserRepo{Db: db}
}

func (ur *UserRepo) FindAll() ([]models.User, error) {
	db := ur.Db
	rows := make([]models.User, 0)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))
		b.ForEach(func(k, v []byte) error {
			user := models.User{}
			_ = json.Unmarshal(v, &user)
			rows = append(rows, user)
			return nil
		})
		return nil
	})

	return rows, err
}

func (ur *UserRepo) FindOne(id int) (*models.User, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	db := ur.Db
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))
		v := b.Get(idb)

		if v == nil {
			return fmt.Errorf("Data tidak ditemukan.")
		}

		err := json.Unmarshal(v, user)
		if err != nil {
			return err
		}
		return nil
	})

	return user, err
}

func (ur *UserRepo) Create(user *models.User) (*models.User, error) {
	db := ur.Db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))

		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		user.ID = int64(id)
		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}

		bId := new(bytes.Buffer)
		binary.Write(bId, binary.LittleEndian, id)

		return b.Put(bId.Bytes(), buf)
	})
	return user, err
}

func (ur *UserRepo) Update(id int, user *models.User) (*models.User, error) {
	idb, err := utils.IntToBytes(int64(id))
	if err != nil {
		return user, err
	}

	err = ur.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))
		c := b.Cursor()

		k, v := c.Seek(idb)
		if k == nil {
			return fmt.Errorf("id %d tidak ditemukan.", id)
		}
		err = json.Unmarshal(v, user)

		return nil
	})
	return user, nil
}

func (ur *UserRepo) Delete(id int) error {
	db := ur.Db
	bId, err := utils.IntToBytes(int64(id))
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("user"))
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
