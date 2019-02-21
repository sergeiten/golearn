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

const (
	usersCollection      = "users"
	statesCollection     = "states"
	wordsCollection      = "words"
	activitiesCollection = "stats"
)

// Service of mongodb
type Service struct {
	session *mgo.Session
	db      string
}

// New returns new instance of mockService
func New(cfg *golearn.Config) (*Service, error) {
	session, err := newSession(cfg)
	if err != nil {
		return nil, err
	}

	return &Service{
		session: session,
		db:      cfg.Database.Name,
	}, nil
}

func newSession(cfg *golearn.Config) (*mgo.Session, error) {
	connString := fmt.Sprintf("%s:%s", cfg.Database.Host, cfg.Database.Port)

	if cfg.Database.Delay > 0 {
		time.Sleep(time.Duration(cfg.Database.Delay) * time.Second)
	}

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
func (s Service) RandomQuestion(category string) (golearn.Row, error) {
	r := golearn.Row{}

	condition := bson.M{}

	if category != "" {
		condition = bson.M{
			"category": category,
		}
	}

	count, err := s.session.DB(s.db).C(wordsCollection).Find(condition).Count()
	if err != nil {
		return r, err
	}

	query := s.session.DB(s.db).C(wordsCollection).Find(condition).Limit(1).Skip(rand.Intn(count))

	err = query.One(&r)

	return r, err
}

// RandomAnswers returns random answers with passed count of total answers.
// q is right answer which will be appended to result slice and will be shuffled later.
// count is total amount of answers which is included the right one.
//
// In order to get random rows function generates random number which is used as offset.
// In order to guarantee that function returns exactly passed "count" of words
// function calculated maximum allowed random number. If generated random number of greater
// that allowed function uses maximum number.
//
// Example:
// In case if passed count is equal 4. Function has to find 3 random rows and 1 (the passed one)
// will be appended to result slice.
// Generated random number is equal 7, that means function returns just 1 rows (8).
// Function moves cursor to position 5 to return enough rows (6,7,8).
/*
                                  Moved         Generated
                                    +---------------+
                                    |               |
                                    v               v
+-------+-------+-------+-------+---+---+-------+---+---+-------+
|       |       |       |       |       |       |       |       |
|   1   |   2   |   3   |   4   |   5   |   6   |   7   |   8   |
|       |       |       |       |       |       |       |       |
+-------+-------+-------+-------+-------+-------+-------+-------+
*/
func (s Service) RandomAnswers(q golearn.Row, count int) ([]golearn.Row, error) {
	var r []golearn.Row

	f := bson.M{
		"word": bson.M{
			"$ne": q.Word,
		},
	}

	total, err := s.session.DB(s.db).C(wordsCollection).Find(f).Count()
	if err != nil {
		return r, err
	}

	maxRand := (total - count) + 1

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(total)

	if random > maxRand {
		random = maxRand
	}

	err = s.session.DB(s.db).C(wordsCollection).Find(f).Limit(count - 1).Skip(random).All(&r)
	if err != nil {
		return r, err
	}

	r = append(r, q)

	return r, nil
}

// SetState save latest given set of question and answers
func (s Service) SetState(state golearn.State) error {
	return s.session.DB(s.db).C(statesCollection).Insert(state)
}

// GetState returns lastest saved user state
func (s Service) GetState(userKey string) (golearn.State, error) {
	state := golearn.State{}
	err := s.session.DB(s.db).C(statesCollection).Find(bson.M{"userkey": userKey}).Sort("-timestamp").One(&state)

	return state, err
}

// ResetState resets user state
func (s Service) ResetState(userKey string) error {
	return nil
}

// InsertWord inserts new row to words collection
func (s Service) InsertWord(w golearn.Row) error {
	return s.session.DB(s.db).C(wordsCollection).Insert(w)
}

// InsertUser inserts new user to users collection
func (s Service) InsertUser(user golearn.User) error {
	return s.session.DB(s.db).C(usersCollection).Insert(user)
}

// UpdateUser updates user
func (s Service) UpdateUser(user golearn.User) error {
	return s.session.DB(s.db).C(usersCollection).Update(bson.M{"userid": user.UserID}, user)
}

// ExistUser returns bool if user already exists in db
func (s Service) ExistUser(user golearn.User) (bool, error) {
	if user.UserID == "" {
		return false, errors.New("passed user has empty id")
	}
	count, err := s.session.DB(s.db).C(usersCollection).Find(bson.M{"userid": user.UserID}).Count()
	return count > 0, err
}

// GetUser returns user from db
func (s Service) GetUser(userID string) (golearn.User, error) {
	u := golearn.User{}
	if userID == "" {
		return u, errors.New("passed user id is empty")
	}
	err := s.session.DB(s.db).C(usersCollection).Find(bson.M{"userid": userID}).One(&u)

	return u, err
}

// SetUserMode sets new mode for passed user id
func (s Service) SetUserMode(userID string, mode string) error {
	return s.session.DB(s.db).C(usersCollection).Update(bson.M{"userid": userID}, bson.M{
		"$set": bson.M{
			"mode": mode,
		},
	})
}

// GetCategories returns list of unique categories based on words table.
func (s Service) GetCategories(userID string) ([]golearn.Category, error) {
	var categories []golearn.Category

	err := s.session.DB(s.db).C(wordsCollection).Pipe([]bson.M{
		{
			"$match": bson.M{
				"category": bson.M{
					"$ne": "",
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$category",
				"name": bson.M{
					"$first": "$category",
				},
				"words": bson.M{
					"$sum": 1,
				},
				"category": bson.M{
					"$first": "$category",
				},
			},
		},
		{
			"$sort": bson.M{
				"category": 1,
			},
		},
	}).All(&categories)

	return categories, err
}

func (s Service) SetUserCategory(userID string, category string) error {
	return s.session.DB(s.db).C(usersCollection).Update(bson.M{"userid": userID}, bson.M{
		"$set": bson.M{
			"category": category,
		},
	})
}

func (s Service) InsertActivity(activity golearn.Activity) error {
	return s.session.DB(s.db).C(activitiesCollection).Insert(activity)
}
