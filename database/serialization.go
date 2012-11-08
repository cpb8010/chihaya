package database

import (
	"chihaya/config"
	"encoding/gob"
	"log"
	"os"
	"time"
)

func (db *Database) startSerializing() {
	go func() {
		for !db.terminate {
			time.Sleep(config.DatabaseSerializationInterval)
			db.serialize()
		}
	}()
}

func (db *Database) serialize() {
	torrentFile, err := os.OpenFile("torrent-cache.gob", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("!!! CRITICAL !!! Couldn't open torrent cache file for writing! ", err)
		return
	}

	userFile, err := os.OpenFile("user-cache.gob", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println("!!! CRITICAL !!! Couldn't open user cache file for writing! ", err)
		return
	}

	defer torrentFile.Close()
	defer userFile.Close()

	log.Printf("Serializing database to cache file")

	db.TorrentsMutex.RLock()
	gob.NewEncoder(torrentFile).Encode(db.Torrents)
	db.TorrentsMutex.RUnlock()

	db.UsersMutex.RLock()
	gob.NewEncoder(userFile).Encode(db.Users)
	db.UsersMutex.RUnlock()
}

func (db *Database) deserialize() {
	torrentFile, err := os.OpenFile("torrent-cache.gob", os.O_RDONLY, 0600)
	if err != nil {
		log.Println("Torrent cache missing, skipping deserialization")
		return
	}
	userFile, err := os.OpenFile("user-cache.gob", os.O_RDONLY, 0600)
	if err != nil {
		log.Println("User cache missing, skipping deserialization")
		return
	}

	defer torrentFile.Close()
	defer userFile.Close()

	log.Printf("Deserializing database from cache file")

	decoder := gob.NewDecoder(torrentFile)

	db.TorrentsMutex.Lock()
	err = decoder.Decode(&db.Torrents)
	db.TorrentsMutex.Unlock()

	if err != nil {
		log.Println("!!! CRITICAL !!! Failed to deserialize torrent cache! You may need to delete it.", err)
		panic("Torrent deserialization failed")
	}

	decoder = gob.NewDecoder(userFile)

	db.UsersMutex.Lock()
	err = decoder.Decode(&db.Users)
	db.UsersMutex.Unlock()

	if err != nil {
		log.Println("!!! CRITICAL !!! Failed to deserialize user cache! You may need to delete it.", err)
		panic("Torrent deserialization failed")
	}

	db.TorrentsMutex.RLock()
	peers := 0
	torrents := len(db.Torrents)
	for _, t := range db.Torrents {
		peers += len(t.Leechers) + len(t.Seeders)
	}
	db.TorrentsMutex.RUnlock()

	db.UsersMutex.RLock()
	users := len(db.Users)
	db.UsersMutex.RUnlock()

	log.Printf("Loaded %d users, %d torrents, %d peers\n", users, torrents, peers)
}