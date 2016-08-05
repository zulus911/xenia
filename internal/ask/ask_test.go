package ask_test

import (
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/tests"

	"github.com/coralproject/xenia/internal/ask"
	"github.com/coralproject/xenia/internal/ask/afix"
	"github.com/coralproject/xenia/internal/form"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {

	var (
		db *db.DB
	)

	BeforeEach(func() {
		// get a database from the fixtures
		db = afix.SetTestDatabase()
	})

	AfterEach(func() {
		// empty database
		afix.TearTestDatabase(db)
	})

	Describe("a form", func() {

		var (
			fixtureForm *form.Form
			result      *form.Form
			err         error
		)

		JustBeforeEach(func() {

			tests.ResetLog()
			defer tests.DisplayLog()

			if fixtureForm, err = afix.GetFixtureForm(); err != nil {
				Expect(err).Should(BeNil(), "Not able to load fixture form.")
			}

			result, err = ask.UpsertForm(tests.Context, db, fixtureForm)
		})

		Context("with validated data", func() {

			It("should not give an error", func() {
				Expect(err).Should(BeNil())
			})

			It("should save its description", func() {
				expectedDescription := "of the rest of your life"
				Expect(result.Header.(map[string]interface{})["description"]).Should(Equal(expectedDescription))
			})
		})
	})
})

var _ = Describe("Update", func() {

	var (
		db *db.DB
	)

	BeforeEach(func() {
		// get a database from the fixtures
		db = afix.SetTestDatabase()
	})

	AfterEach(func() {
		// empty database
		afix.TearTestDatabase(db)
	})

	Describe("a form", func() {

		var (
			fixtureForm *form.Form
			result      *form.Form
			err         error
		)

		JustBeforeEach(func() {

			tests.ResetLog()
			defer tests.DisplayLog()

			// load the fixture form
			if fixtureForm, err = afix.GetFixtureForm(); err != nil {
				Expect(err).Should(BeNil(), "Not able to load fixture form.")
			}

			// insert the form into the coral database
			if _, err = ask.UpsertForm(tests.Context, db, fixtureForm); err != nil {
				Expect(err).Should(BeNil(), "Not able to create fixture form.")
			}

			fixtureForm.Status = "new"
			result, err = ask.UpsertForm(tests.Context, db, fixtureForm)
		})

		Context("with validated data", func() {

			It("should not give an error", func() {
				Expect(err).Should(BeNil())
			})

			It("should save its status", func() {

				expectedStatus := "new"
				Expect(result.Status).Should(Equal(expectedStatus))

			})
		})
	})

	Describe("a form submission", func() {

		var (
			submission *form.Submission
			errSub     error
		)

		JustBeforeEach(func() {

			tests.ResetLog()
			defer tests.DisplayLog()

			// load the fixture form
			fixtureForm, _ := afix.GetFixtureForm()

			// insert the form into the coral database
			_, _ = ask.UpsertForm(tests.Context, db, fixtureForm)

			// load the fixture form
			fixtureFormSubmissionInput, err := afix.GetFixtureFormSubmissionInput()
			Expect(err).Should(BeNil(), "Not able to load fixture form submission.")

			// insert the form into the coral database
			submission, errSub = ask.UpsertFormSubmission(tests.Context, db, fixtureFormSubmissionInput)
		})

		Context("with appropiate context", func() {
			It("should not give an error", func() {
				Expect(errSub).Should(BeNil())
				Expect(submission).ShouldNot(BeNil())
			})
			It("should return a submission for its form", func() {
				expectedFormID := "577c18f4a969c805f7f8c889"
				Expect(submission.FormID.Hex()).Should(Equal(expectedFormID))
			})
		})
	})

})
