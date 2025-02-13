package db

import (
	"encoding/json"
	"fmt"
	"log"

	"go.etcd.io/bbolt"
)

var db *bbolt.DB

// InitDB initializes the database
func InitDB() {
	var err error
	db, err = bbolt.Open("data.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create "items" bucket if not exists
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("items"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

// SetItem stores a key-value pair where the value can be of any type
func SetItem(key string, value interface{}) error {
	// Serialize value to JSON
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %v", err)
	}

	// Store in BoltDB
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Put([]byte(key), valueBytes)
	})
	if err != nil {
		return fmt.Errorf("failed to save data: %v", err)
	}
	return nil
}

// GetItem retrieves a value by key and deserializes it into a provided variable
func GetItem(key string, result interface{}) error {
	// Retrieve value from BoltDB
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		v := b.Get([]byte(key))
		if v == nil {
			return fmt.Errorf("key not found")
		}
		return json.Unmarshal(v, result) // Deserialize JSON into result
	})
	if err != nil {
		return fmt.Errorf("error retrieving key: %v", err)
	}

	return nil
}

// DeleteItem deletes a key-value pair
func DeleteItem(key string) error {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("failed to delete key: %v", err)
	}
	return nil
}
