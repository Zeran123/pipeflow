package bot

import (
	"bytes"
	"encoding/gob"
	"log"

	bolt "go.etcd.io/bbolt"
)

type Db interface {
	save(s Store) (Store, error)
	list() ([]Store, error)
	get(id string) (Store, error)
	del(id string) error
}

type Blot struct {
	filePath   string
	bucketName string
}

func (b Blot) save(s Store) (Store, error) {
	db, err := bolt.Open(b.filePath, 0600, nil)
	if err != nil {
		return s, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		s.Id = RandStr(36)
		b, _ := tx.CreateBucketIfNotExists([]byte(b.bucketName))
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(s)
		err = b.Put([]byte(s.Id), buf.Bytes())
		return err
	})
	defer db.Close()
	if err != nil {
		return s, err
	}
	return s, nil
}

func (b Blot) list() ([]Store, error) {
	ss := make([]Store, 0, 5)
	db, err := bolt.Open(b.filePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var s Store
			buf := bytes.NewBuffer(v)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&s)
			if err != nil {
				log.Fatal(err)
			}
			ss = append(ss, s)
		}
		return err
	})
	defer db.Close()
	if err != nil {
		return nil, err
	}
	return ss, nil
}

func (b Blot) get(id string) (Store, error) {
	var s Store
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return s, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucket))
		v := b.Get([]byte(id))
		buf := bytes.NewBuffer(v)
		enc := gob.NewDecoder(buf)
		err := enc.Decode(&s)
		return err
	})
	defer db.Close()
	if err != nil {
		return s, err
	}
	return s, nil
}

func (b Blot) del(id string) error {
	db, err := bolt.Open(b.filePath, 0600, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucketName))
		err := b.Delete([]byte(id))
		return err
	})
	defer db.Close()
	return nil
}
