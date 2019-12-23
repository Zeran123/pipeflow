package main

import (
	"bytes"
	"encoding/gob"
	"log"

	bolt "go.etcd.io/bbolt"
)

type Db interface {
	save(bot Bot) (Bot, error)
	list() ([]Bot, error)
	get(id string) (Bot, error)
	del(id string) error
}

type Blot struct {
	filePath   string
	bucketName string
}

func (b Blot) save(bot Bot) (Bot, error) {
	db, err := bolt.Open(b.filePath, 0600, nil)
	if err != nil {
		return bot, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bot.Id = RandStr(36)
		b, _ := tx.CreateBucketIfNotExists([]byte(b.bucketName))
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(bot)
		err = b.Put([]byte(bot.Id), buf.Bytes())
		return err
	})
	defer db.Close()
	if err != nil {
		return bot, err
	}
	return bot, nil
}

func (b Blot) list() ([]Bot, error) {
	bots := make([]Bot, 0, 5)
	db, err := bolt.Open(b.filePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b.bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var bot Bot
			buf := bytes.NewBuffer(v)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&bot)
			if err != nil {
				log.Fatal(err)
			}
			bots = append(bots, bot)
		}
		return err
	})
	defer db.Close()
	if err != nil {
		return nil, err
	}
	return bots, nil
}

func (b Blot) get(id string) (Bot, error) {
	var bot Bot
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return bot, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucket))
		v := b.Get([]byte(id))
		buf := bytes.NewBuffer(v)
		enc := gob.NewDecoder(buf)
		err := enc.Decode(&bot)
		return err
	})
	defer db.Close()
	if err != nil {
		return bot, err
	}
	return bot, nil
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
