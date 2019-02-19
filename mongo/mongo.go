package mongo

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sergeiten/golearn"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Service of mongodb
type Service struct {
	session *mgo.Session
}

// New returns new instance of mockService
func New(cfg *golearn.Config) (*Service, error) {
	session, err := newSession(cfg)
	if err != nil {
		return nil, err
	}

	return &Service{
		session: session,
	}, nil
}

func newSession(cfg *golearn.Config) (*mgo.Session, error) {
	connString := fmt.Sprintf("%s:%s", cfg.Database.Host, cfg.Database.Port)

	timeout := time.Duration(5 * time.Second)
	session, err := mgo.DialWithTimeout(connString, timeout)

	if err != nil {
		return session, err
	}
	session.SetMode(mgo.Monotonic, true)

	return session, nil
}

// Close terminates the mockService session
func (s Service) Close() {
	s.session.Close()
}

// RandomQuestion returns random row
func (s Service) RandomQuestion() (golearn.Row, error) {
	r := golearn.Row{}

	count, err := s.session.DB("golearn").C("words").Count()
	if err != nil {
		return r, err
	}

	query := s.session.DB("golearn").C("words").Find(bson.M{}).Limit(1).Skip(rand.Intn(count))

	err = query.One(&r)

	return r, err
}

// RandomAnswers returns random answers with given limit
func (s Service) RandomAnswers(q golearn.Row, limit int) ([]golearn.Row, error) {
	r := []golearn.Row{}

	count, err := s.session.DB("golearn").C("words").Count()
	if err != nil {
		return r, err
	}

	f := bson.M{
		"word": bson.M{
			"$ne": q.Word,
		},
	}

	err = s.session.DB("golearn").C("words").Find(f).Limit(3).Skip(rand.Intn(count)).All(&r)
	if err != nil {
		return r, err
	}

	r = append(r, q)

	return r, nil
}

// SetState save latest given set of question and answers
func (s Service) SetState(state golearn.State) error {
	return s.session.DB("golearn").C("states").Insert(state)
}

// GetState returns lastest saved user state
func (s Service) GetState(userKey string) (golearn.State, error) {
	state := golearn.State{}
	err := s.session.DB("golearn").C("states").Find(bson.M{"userkey": userKey}).Sort("-timestamp").One(&state)

	return state, err
}

// ResetState resets user state
func (s Service) ResetState(userKey string) error {
	return nil
}

// InsertWord inserts new row to words collection
func (s Service) InsertWord(w golearn.Row) error {
	return s.session.DB("golearn").C("words").Insert(w)
}

// InsertUser inserts new user to users collection
func (s Service) InsertUser(user golearn.User) error {
	return s.session.DB("golearn").C("users").Insert(user)
}

// UpdateUser updates user
func (s Service) UpdateUser(user golearn.User) error {
	return s.session.DB("golearn").C("users").Update(bson.M{"userid": user.UserID}, user)
}

// ExistUser returns bool if user already exists in db
func (s Service) ExistUser(user golearn.User) (bool, error) {
	if user.UserID == "" {
		return false, errors.New("passed user has empty id")
	}
	count, err := s.session.DB("golearn").C("users").Find(bson.M{"userid": user.UserID}).Count()
	return count > 0, err
}

// GetUser returns user from db
func (s Service) GetUser(userID string) (golearn.User, error) {
	u := golearn.User{}
	if userID == "" {
		return u, errors.New("passed user id is empty")
	}
	err := s.session.DB("golearn").C("users").Find(bson.M{"userid": userID}).One(&u)

	return u, err
}

// SetUserMode sets new mode for passed user id
func (s Service) SetUserMode(userID string, mode string) error {
	return s.session.DB("golearn").C("users").Update(bson.M{"userid": userID}, bson.M{
		"$set": bson.M{
			"mode": mode,
		},
	})
}

// GetCategories returns list of unique categories based on words table.
func (s Service) GetCategories(userID string) ([]golearn.Category, error) {
	var categories []golearn.Category

	err := s.session.DB("golearn").C("words").Pipe(bson.M{
		"$group": bson.M{
			"_id": "$category",
			"name": bson.M{
				"$first": "$category",
			},
			"words": bson.M{
				"$sum": 1,
			},
		},
	}).All(&categories)

	return categories, err
}

func (s Service) SetUserCategory(userID string, category string) error {
	return s.session.DB("golearn").C("users").Update(bson.M{"userid": userID}, bson.M{
		"$set": bson.M{
			"category": category,
		},
	})
}
