package ask

import (
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/log"
	"github.com/coralproject/xenia/internal/form"
)

// UpsertForm creates or update an existing form
func UpsertForm(context interface{}, db *db.DB, f *form.Form) (*form.Form, error) {
	log.Dev(context, "Upsert", "Started : Identifier[%s]", f.Id())

	rf, err := form.UpsertForm(context, db, f)
	if err != nil {
		return nil, err
	}

	log.Dev(context, "Upsert", "Completed")
	return rf, nil
}

func UpsertFormSubmission(context interface{}, db *db.DB, f *form.SubmissionInput) (*form.Submission, error) {
	log.Dev(context, "Upsert", "Started : Identifier[%s]", f.Id())

	rs, err := form.UpsertSubmission(context, db, f)
	if err != nil {
		return nil, err
	}

	log.Dev(context, "Upsert", "Completed")
	return rs, nil
}
