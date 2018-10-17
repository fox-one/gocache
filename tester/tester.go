package tester

import (
	"fmt"
	"testing"

	"github.com/fox-one/gocache"
	"github.com/stretchr/testify/assert"
)

type User struct {
	Id    uint   `json:"id"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (u User) CacheKey() (string, error) {
	if u.Id > 0 {
		return fmt.Sprintf("cache_user_id_%d", u.Id), nil
	}

	return "", fmt.Errorf("id is invalid")
}

func (u User) CacheExpire() int64 {
	return 60 * 60 // 1 hour
}

func (u User) CacheSubkeys() []string {
	keys := []string{}

	if len(u.Email) > 0 {
		keys = append(keys, "cache_user_email_"+u.Email)
	}

	if len(u.Phone) > 0 {
		keys = append(keys, "cache_user_tel_"+u.Phone)
	}

	return keys
}

func Save(t *testing.T, store gocache.Store, coder gocache.Coder) {
	c := gocache.New(store, coder)

	u := User{
		Id:    1,
		Phone: "18512345678",
		Email: "test@fox.one",
		Name:  "tom",
	}

	pk, _ := u.CacheKey()
	subkeys := u.CacheSubkeys()
	keys := append(subkeys, pk)

	if ok := assert.Nil(t, c.Clean(u)); !ok {
		return
	}

	for _, key := range keys {
		exist, err := store.Exists(key)
		if !assert.Nil(t, err) || !assert.False(t, exist) {
			return
		}
	}

	userWithoutEmail := u
	userWithoutEmail.Email = ""

	if err := c.Save(userWithoutEmail); !assert.Nil(t, err) {
		return
	}

	if err := c.Load(&User{Email: u.Email}); !assert.Equal(t, gocache.CacheMiss, err) {
		return
	}

	// load cache by primary key
	userWithId := User{Id: u.Id}
	if err := c.Load(&userWithId); !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, userWithoutEmail, userWithId) {
		return
	}

	// load cache by subkey
	userWithPhone := User{Phone: u.Phone}
	if err := c.Load(&userWithPhone); !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, userWithoutEmail, userWithPhone) {
		return
	}

	c.Save(u)
	userWithoutId := u
	userWithoutId.Id = 0
	c.Load(&userWithoutId)
	assert.Equal(t, u, userWithoutId)
}
