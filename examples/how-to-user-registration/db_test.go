package main

import "testing"

func TestUserDB(t *testing.T) {
	db := userDB{
		fname: "/tmp/udb.json",
	}

	users := []string{
		"first@test.com",
		"second@test.com",
		"third@test.com",
	}

	for _, ustr := range users {
		u, err := db.createUser(ustr, "aaa"+ustr[:5])
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("user created: %v", u)
	}

	db = loadDB(db.fname)

	for _, ustr := range users {
		u, found := db.findUser(ustr)
		if found {
			t.Logf("user found: %v", u)
		} else {
			t.Fatal("user not found", ustr)
		}
	}
}
