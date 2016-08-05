package form

import (
	"fmt"
	"time"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Gallery implements the Model interface
type Gallery struct {
	ID          bson.ObjectId          `json:"id" bson:"_id"`
	FormID      bson.ObjectId          `json:"form_id" bson:"form_id"`
	Headline    string                 `json:"headline" bson:"headline"`
	Description string                 `json:"description" bson:"description"`
	Config      map[string]interface{} `json:"config" bson:"config"`
	Answers     []GalleryAnswer        `json:"answers" bson:"answers"`
	DateCreated time.Time              `json:"date_created,omitempty" bson:"date_created,omitempty"`
	DateUpdated time.Time              `json:"date_updated,omitempty" bson:"date_updated,omitempty"`
}

type GalleryAnswer struct {
	SubmissionID    bson.ObjectId      `json:"submission_id" bson:"submission_id"`
	AnswerID        string             `json:"answer_id" bson:"answer_id"`
	Answer          SubmissionAnswer   `json:"answer,omitempty" bson:"answer,omitempty"`                     // not saved to db, hydrated when reading only!
	IdentityAnswers []SubmissionAnswer `json:"identity_answers,omitempty" bson:"identity_answers,omitempty"` // not saved to db, hydrated when reading only!
}

// Gallery I am, I am form_gallery
func (o Gallery) GetType() string {
	return "form_gallery"
}

// IsRecordableEvent record all Historical Events for Galleries
func (o Gallery) IsRecordableEvent(e string) bool {
	return true
}

// Id returns the ID for this Model
func (object Gallery) Id() string {
	return object.ID.Hex()
}

func (object Gallery) Validate() error {
	errs := validate.Struct(object)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

//=========================================================================================================

func Create(context interface{}, db *db.DB, formID string) (*Gallery, error) {

	// create a new gallery and set it up
	fg := Gallery{
		FormID:      bson.ObjectIdHex(formID),
		DateCreated: time.Now(),
		DateUpdated: time.Now(),
	}

	fg.ID = bson.NewObjectId()

	f := func(c *mgo.Collection) error {
		return c.Insert(&fg)
	}

	if err := db.ExecuteMGO(context, FormGalleries, f); err != nil {
		log.Error(context, "Upsert", err, "Completed")
		return nil, err
	}

	// THIS CAN BE ACHIEVED IN A STRUCTURED LOG
	// // store the history of it's creation!
	// hr := behavior.HistoricalRecord{}
	// hr.Record("Created", fg)

	return &fg, nil

}

// embeds the latest version of the FormSubmisison.Answer into
//  a Form Gallery.  Loaded every time to react to Edits/deltes
//  of form submission content
//
// Identity is defined by answers to form questions that are tagged with
//   identity: true. In addition to capturing the answers, this func stores
//  all the identity information for each submission
// func hydrateFormGallery(g model.FormGallery) model.FormGallery {
//
// 	// get a context to load the submissions
// 	c := web.NewContext(nil, nil)
//
// 	// for each answer in the gallery
// 	for i, a := range g.Answers {
//
// 		// load the submission
// 		c.SetValue("id", a.SubmissionId.Hex())
// 		s, err := GetFormSubmission(c)
// 		if err != nil {
// 			// remove answers from gallery if submission is
// 			//  deleted?
// 		}
//
// 		// find the answer
// 		for _, fsa := range s.Answers {
// 			if fsa.WidgetId == a.AnswerId {
//
// 				// and embed it into the form gallery
// 				g.Answers[i].Answer = fsa
//
// 				// now let's package up the identity flagged answers
// 				// create a slice of answers to contain identity fields
// 				g.Answers[i].IdentityAnswers = []model.FormSubmissionAnswer{}
//
// 				for _, ifsa := range s.Answers {
//
// 					// append all answers flagged as identity to this answer
// 					if ifsa.Identity == true {
// 						//						fmt.Println("found identity!", i, ifsa)
// 						g.Answers[i].IdentityAnswers = append(g.Answers[i].IdentityAnswers, ifsa)
// 					}
// 				}
//
// 			}
//
// 		}
// 	}
//
// 	return g
// }
