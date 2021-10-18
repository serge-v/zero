package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"os"
)

type user struct {
	Email      string
	Password   string
	Confirmed  bool
	RandomHash string
}

type userDB struct {
	fname string
	Users []user
}

func loadDB(fname string) userDB {
	db := userDB{fname: fname}
	f, err := os.Open(fname)
	if err != nil && os.IsNotExist(err) {
		return db
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&db); err != nil {
		log.Fatal(err)
	}

	return db
}

func (db *userDB) saveDB() error {
	f, err := os.Create(db.fname)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(db); err != nil {
		return err
	}
	return nil
}

func (db *userDB) findUser(email string) (*user, bool) {
	for _, u := range db.Users {
		if u.Email == email {
			return &u, true
		}
	}
	return nil, false
}

func (db *userDB) createUser(email, password string) (*user, error) {
	if u, found := db.findUser(email); found {
		return u, nil
	}

	randomHash, err := generateRandomString(20)
	if err != nil {
		return nil, err
	}

	u := user{
		Email:      email,
		Password:   password,
		RandomHash: randomHash,
	}

	db.Users = append(db.Users, u)
	if err := db.saveDB(); err != nil {
		return nil, err
	}

	return &u, nil
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
