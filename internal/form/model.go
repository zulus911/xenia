package form

import "gopkg.in/bluesuncorp/validator.v8"

//Various Constants
const (
	// Ask collections
	Forms           string = "forms"
	FormSubmissions string = "form_submissions"
	FormGalleries   string = "form_galleries"
)

// validate is used to perform model field validation.
var validate *validator.Validate

func init() {
	validate = validator.New(&validator.Config{TagName: "validate"})
}

type Model interface {
	Id() string
	Validate() error
}
