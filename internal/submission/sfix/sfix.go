package sfix

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/mgo.v2"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/tests"
	"github.com/coralproject/xenia/internal/form"
	"github.com/coralproject/xenia/internal/submission"
)

const (
	dataForm                = "form.json"
	dataFormSubmission      = "submission.json"
	dataFormSubmissionInput = "submission_input.json"
)

var (
	path string
)

func init() {
	path = os.Getenv("GOPATH") + "/src/github.com/coralproject/xenia/internal/submission/sfix/"
}

// SetTestDatabase in the tests we are creating a unique database for testing
// and removing it after running all the tests
func SetTestDatabase() *db.DB {

	// In order to get a Mongo session we need the name of the database we
	// are using. The web framework middleware is using this by convention.
	db, err := db.NewMGO(tests.Context, tests.TestSession)
	if err != nil {
		fmt.Println("Unable to get Mongo session")
		return nil
	}

	return db
}

// remove the test database
func TearTestDatabase(db *db.DB) {

	collections := []string{submission.FormSubmissions}

	f := func(c *mgo.Collection) error {
		return c.DropCollection()
	}

	for _, c := range collections {
		_ = db.ExecuteMGO(tests.Context, c, f)
	}

	db.CloseMGO(tests.Context)
}

/*==================================================================================*/

// get a fixture from appropiate json file for data form
func GetFixtureForm() (*form.Form, error) {

	file, err := os.Open(path + dataForm)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var form *form.Form
	err = json.NewDecoder(file).Decode(&form)

	return form, err
}

// get a fixture for the form submission
func GetFixtureFormSubmission() (*submission.Submission, error) {

	file, err := os.Open(path + dataFormSubmission)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var submission *submission.Submission
	err = json.NewDecoder(file).Decode(&submission)

	return submission, err
}

// get a fixtuer for the submission input
func GetFixtureFormSubmissionInput() (*submission.SubmissionInput, error) {

	file, err := os.Open(path + dataFormSubmissionInput)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var submission *submission.SubmissionInput
	err = json.NewDecoder(file).Decode(&submission)

	return submission, err
}
