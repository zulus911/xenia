package form

import (
	"fmt"
	"time"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/db/mongo"
	"github.com/ardanlabs/kit/log"

	"gopkg.in/bluesuncorp/validator.v8"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Various Constants
const (
	// Ask collections
	Forms           string = "forms"
	FormSubmissions string = "form_submissions"
)

// validate is used to perform model field validation.
var validate *validator.Validate

func init() {
	validate = validator.New(&validator.Config{TagName: "validate"})
}

// Form is our main structure for the form builder
// It implements the interface Model
type Form struct {
	ID             bson.ObjectId `json:"id" bson:"_id"`
	Status         string        `json:"status" bson:"status"`
	Theme          interface{}   `json:"theme" bson:"theme"`
	Settings       interface{}   `json:"settings" bson:"settings"`
	Header         interface{}   `json:"header" bson:"header"`
	Footer         interface{}   `json:"footer" bson:"footer"`
	FinishedScreen interface{}   `json:"finishedScreen" bson:"finishedScreen"`
	Steps          []FormStep    `json:"steps" bson:"steps"`
	Stats          FormStats     `json:"stats" bson:"stats"`
	CreatedBy      interface{}   `json:"created_by" bson:"created_by"` // Todo, decide how to represent ownership here
	UpdatedBy      interface{}   `json:"updated_by" bson:"updated_by"` // Todo, decide how to represent ownership here
	DeletedBy      interface{}   `json:"deleted_by" bson:"deleted_by"` // Todo, decide how to represent ownership here
	DateCreated    time.Time     `json:"date_created,omitempty" bson:"date_created,omitempty"`
	DateUpdated    time.Time     `json:"date_updated,omitempty" bson:"date_updated,omitempty"`
	DateDeleted    time.Time     `json:"date_deleted,omitempty" bson:"date_deleted,omitempty"`
}

type FormStats struct {
	Responses int `json:"responses" bson:"responses"`
}

type FormStep struct {
	ID      string       `json:"id" bson:"_id"`
	Name    string       `json:"name" bson:"name"`
	Widgets []FormWidget `json:"widgets" bson:"widgets"`
}

type FormWidget struct {
	ID          string      `json:"id" bson:"_id"`
	Type        string      `json:"type" bson:"type"`
	Identity    bool        `json:"identity" bson:"identity"`
	Component   string      `json:"component" bson:"component"`
	Title       string      `json:"title" bson:"title"`
	Description string      `json:"description" bson:"description"`
	Wrapper     interface{} `json:"wrapper" bson:"wrapper"`
	Props       interface{} `json:"props" bson:"props"`
}

// Id returns the ID for this Model
func (object Form) Id() string {
	return object.ID.Hex()
}

func (object Form) Validate() error {
	errs := validate.Struct(object)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

func (object Form) CountSubmissions(context interface{}, db *db.DB) (int, error) {

	var n int
	var err error

	f := func(c *mgo.Collection) error {
		n, err = c.Find(bson.M{"form_id": object.ID}).Count()
		return err
	}

	if err := db.ExecuteMGO(context, FormSubmissions, f); err != nil {
		log.Error(context, "CountSubmissions", err, "Completed")
		return 0, err
	}

	return n, err
}

// calculate stats for Forms
func (object Form) UpdateStats(context interface{}, db *db.DB) error {

	// do some counting
	responses, err := object.CountSubmissions(context, db)
	if err != nil {
		return err
	}

	// update the stats subdoc

	s := FormStats{}
	s.Responses = responses

	f := func(c *mgo.Collection) error {
		err := c.Update(bson.M{"_id": object.ID}, bson.M{"$set": bson.M{"stats": s}})
		return err
	}

	if err := db.ExecuteMGO(context, Forms, f); err != nil {
		log.Error(context, "updateStats", err, "Completed")
		return err
	}

	return nil

}

// Returns a form based on the form_id
func GetForm(context interface{}, db *db.DB, fID bson.ObjectId) (*Form, error) {

	var f *Form

	funct := func(c *mgo.Collection) error {
		return c.FindId(fID).One(&f)
	}

	if err := db.ExecuteMGO(context, Forms, funct); err != nil {
		return nil, err
	}

	return f, nil
}

//=========================================================================================================

// Upsert create or update an existing form
func Upsert(context interface{}, db *db.DB, f *Form) (*Form, error) {

	// Validate the form that is provided.
	if err := f.Validate(); err != nil {
		log.Error(context, "Upsert", err, "Completed")
		return nil, err
	}

	funct := func(c *mgo.Collection) error {

		var rf *Form
		// Look for the form in the db amd update datetime fields
		if err := c.FindId(f.ID).One(&rf); err != nil {
			//f.ID = bson.NewObjectId()
			f.DateCreated = time.Now()
			return c.Insert(f)
		}
		f.DateUpdated = time.Now()

		// If the form already exists then update it
		q := bson.M{"_id": f.ID}
		log.Dev(context, "Upsert", "MGO : db.%s.upsert(%s, %s)", c.Name, mongo.Query(q), mongo.Query(f))
		return c.Update(q, f)
	}

	if err := db.ExecuteMGO(context, Forms, funct); err != nil {
		log.Error(context, "Upsert", err, "Completed")
		return nil, err
	}

	// always update form stats to ensure expected stats fields
	if err := f.UpdateStats(context, db); err != nil {
		return nil, err
	}

	return f, nil
}
