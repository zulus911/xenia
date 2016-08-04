package form

import (
	"fmt"
	"time"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/db/mongo"
	"github.com/ardanlabs/kit/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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
			f.DateCreated = time.Now()
			return c.Insert(f)
		}

		// If the form already exists then update it
		q := bson.M{"_id": f.ID}
		log.Dev(context, "Upsert", "MGO : db.%s.upsert(%s, %s)", c.Name, mongo.Query(q), mongo.Query(f))
		return c.Update(q, f)
	}

	if err := db.ExecuteMGO(context, Forms, funct); err != nil {
		log.Error(context, "Upsert", err, "Completed")
		return nil, err
	}

	return f, nil

	//
	// // create
	// if input.ID == "" {
	//
	// 	// append a fresh id to the input obj
	// 	input.ID = bson.NewObjectId()
	//
	// 	// and insert it
	// 	if err := c.MDB.DB.C(model.Forms).Insert(input); err != nil {
	// 		return err
	// 	}

	// store the id into the context as a hex
	// //  to match up with what we expect from web params
	// c.SetValue("id", input.ID.Hex())
	//
	// // we're auto-creating galleries for forms
	// //  so create a context and do so
	// fc := web.NewContext(nil, nil)
	// fc := app.
	// defer fc.Close()
	// fc.SetValue("form_id", input.ID.Hex())
	// CreateFormGallery(fc)

	// } else { // do the update
	//
	// 	// store the existing id into the context as a hex
	// 	//  to match up with what we expect from web params
	// 	context.SetValue("id", input.ID.Hex())
	//
	// 	if _, err := c.MDB.DB.C(model.Forms).UpsertId(input.ID, input); err != nil {
	//
	// 		return nil, err
	// 	}

	//}

	// // always update form stats to ensure expected stats fields
	// err := updateStats(context)
	// if err != nil {
	// 	return nil, err
	// }
}
