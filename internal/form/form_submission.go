package form

import (
	"fmt"
	"time"

	"github.com/ardanlabs/kit/db"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Submission implements the Model interface
type Submission struct {
	ID             bson.ObjectId      `json:"id" bson:"_id"`
	FormID         bson.ObjectId      `json:"form_id" bson:"form_id"`
	Number         int                `json:"number" bson:"number"`
	Status         string             `json:"status" bson:"status"`
	Answers        []SubmissionAnswer `json:"replies" bson:"replies"`
	Flags          []string           `json:"flags" bson:"flags"` // simple, flexible string flagging
	Header         interface{}        `json:"header" bson:"header"`
	Footer         interface{}        `json:"footer" bson:"footer"`
	FinishedScreen interface{}        `json:"finishedScreen" bson:"finishedScreen"`
	CreatedBy      interface{}        `json:"created_by" bson:"created_by"` // Todo, decide how to represent ownership here
	UpdatedBy      interface{}        `json:"updated_by" bson:"updated_by"` // Todo, decide how to represent ownership here
	DateCreated    time.Time          `json:"date_created,omitempty" bson:"date_created,omitempty"`
	DateUpdated    time.Time          `json:"date_updated,omitempty" bson:"date_updated,omitempty"`
}

type SubmissionEditInput struct {
	EditedAnswer interface{} `json:"edited"`
}

// this is what we expect for input for a form submission
type SubmissionAnswerInput struct {
	WidgetID string      `json:"widget_id"`
	Answer   interface{} `json:"answer"`
}

// SubmissionInput implements the Model interface
type SubmissionInput struct {
	FormID  string                  `json:"form_id"`
	Status  string                  `json:"status" bson:"status"`
	Answers []SubmissionAnswerInput `json:"replies"`
}

// here's what a form submission is
type SubmissionAnswer struct {
	WidgetID     string      `json:"widget_id" bson:"widget_id"`
	Identity     bool        `json:"identity" bson:"identity"`
	Answer       interface{} `json:"answer" bson:"answer"`
	EditedAnswer interface{} `json:"edited" bson:"edited"`
	Question     interface{} `json:"question" bson:"question"`
	Props        interface{} `json:"props" bson:"props"`
}

// Id returns the ID for this Model
func (object Submission) Id() string {
	return object.ID.Hex()
}

func (object Submission) Validate() error {
	errs := validate.Struct(object)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// Id returns the ID for this Model
func (object SubmissionInput) Id() string {
	return ""
}

func (object SubmissionInput) Validate() error {
	errs := validate.Struct(object)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// it's a little peculiar:
// each submission to a Form will have a record for every answer no
// matter what the fe sends
// these are prepopulated by buildSubmissionFromForm above
// so..
func (object Submission) SetAnswersToSubmission(fsi *SubmissionInput) {

	// for each answer inputted
	for _, ai := range fsi.Answers {

		// look for the answer
		for x, a := range object.Answers {

			// add the answer to the appropriate spot
			if a.WidgetID == ai.WidgetID {
				object.Answers[x].Answer = ai.Answer
			}
		}
	}
}

//=========================================================================================================

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

// Upsert create or update an existing form submission
func UpsertSubmission(context interface{}, db *db.DB, f *SubmissionInput) (*Submission, error) {

	// we need to be sure that the submission input is a valid struct
	if err := f.Validate(); err != nil {
		return nil, err
	}

	// get the form id for the submission
	fID := bson.ObjectIdHex(f.FormID)

	// get the form in question
	form, err := GetForm(context, db, fID)
	if err != nil {
		return nil, err
	}

	// build a form submission from the input
	fs, err := form.BuildSubmission() // buildSubmissionFromForm(f)
	if err != nil {
		return nil, err
	}

	// set the answers into the submission
	fs.SetAnswersToSubmission(f) //fs = setAnswersToFormSubmission(fs, input)

	// set miscellenia
	fs.DateCreated = time.Now()
	fs.DateUpdated = time.Now()

	// set the number
	n, err := form.getSubmissionCountByForm(context, db)
	if err != nil {
		return nil, err
	}
	fs.Number = n + 1

	// aaaand save it
	funct := func(c *mgo.Collection) error {
		return c.Insert(fs)
	}

	if err := db.ExecuteMGO(context, FormSubmissions, funct); err != nil {
		return nil, err
	}

	// // update the stats using the Form Context
	// err = updateStats(fc)
	// if err != nil {
	// 	return nil, err
	// }

	return &fs, nil

}
