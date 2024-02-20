package main

import (
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"

	badger "github.com/dgraph-io/badger/v4"
)

var Badger Database

type Database struct {
	*badger.DB
}

func InitBadger() {
	opts := badger.DefaultOptions("").WithInMemory(true)
	opts.Logger = nil // Disable logging
	var err error
	Badger.DB, err = badger.Open(opts)
	if err != nil {
		log.Fatal().Err(err)
	}
}

// BLOGO SPECIFIC FUNCTIONS

func (d *Database) GetAllArticleSlugs() ([]string, error) {
	var keys []string
	err := d.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("post_")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			keys = append(keys, strings.Trim(string(key), "post_"))
		}
		return nil
	})
	return keys, err
}

func (d *Database) GetPostBySlug(slug string) (ArticleData, error) {
	var value []byte
	var aData ArticleData
	err := d.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("post_" + slug))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})
		return err
	})
	if err != nil {
		return aData, err
	}

	err = json.Unmarshal(value, &aData)
	if err != nil {
		return aData, err
	}
	return aData, err
}

func (d *Database) GetAllArticles() []ArticleData {
	var articles []ArticleData
	articleBytes := d.GetValuesWithPrefix("post_")
	for _, ab := range articleBytes {
		var article ArticleData
		err := json.Unmarshal(ab, &article)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshalling article from Badger:")
			continue
		}
		articles = append(articles, article)
	}
	return articles
}

func (d *Database) DeleteArticle(key string) error {
	return d.Delete("post_" + key)
}

// GENERIC FUNCTIONS

func (d *Database) Set(key string, value []byte) error {
	return d.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (d *Database) Get(key string) ([]byte, error) {
	var value []byte
	err := d.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			value = val
			return nil
		})
		return err
	})
	return value, err
}

func (d *Database) Delete(key string) error {
	return d.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (d *Database) GetAllKeys() ([]string, error) {
	var keys []string
	err := d.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			keys = append(keys, string(key))
		}
		return nil
	})
	return keys, err
}

func (d *Database) GetValues() ([][]byte, error) {
	var values [][]byte
	err := d.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				values = append(values, val)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return values, err
}

func (d *Database) GetValuesWithPrefix(prefix string) [][]byte {
	var values [][]byte
	d.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				values = append(values, val)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return values
}
