package cache

import (
	"fmt"
	"sync"

	"github.com/andygrunwald/go-jira"

	"github.com/legionus/jiramail/internal/jiraplus"
)

type User struct {
	mu     sync.Mutex
	client *jiraplus.Client
	cache  map[string]*jira.User
}

func NewUserCache(c *jiraplus.Client) *User {
	return &User{
		client: c,
		cache:  make(map[string]*jira.User),
	}
}

func (o *User) Set(user *jira.User) {
	if user == nil {
		return
	}

	_, ok := o.cache[user.Name]
	if ok {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.cache[user.Name] = user
}

func (o *User) Get(name string) (*jira.User, error) {
	user, ok := o.cache[name]
	if ok {
		return user, nil
	}

	user, _, err := o.client.User.Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get user %s: %s", name, err)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.cache[name] = user

	return user, nil
}
