package submission_test

import (
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/tests"

	"github.com/coralproject/xenia/internal/form"
	"github.com/coralproject/xenia/internal/submission"
	"github.com/coralproject/xenia/internal/submission/sfix"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {

	var (
		db *db.DB
	)

	BeforeEach(func() {
		// get a database from the fixtures
		db = sfix.SetTestDatabase()
	})

	AfterEach(func() {
		// empty database
		sfix.TearTestDatabase(db)
	})

	Describe("a form submission", func() {

		var (
			s      *submission.Submission
			errSub error
		)

		JustBeforeEach(func() {

			tests.ResetLog()
			defer tests.DisplayLog()

			// load the fixture form
			fixtureForm, err := sfix.GetFixtureForm()
			Expect(err).Should(BeNil(), "Not able to load fixture form.")

			// insert the form into the coral database
			_, _ = form.Upsert(tests.Context, db, fixtureForm)

			// load the fixture form
			fixtureFormSubmissionInput, err := sfix.GetFixtureFormSubmissionInput()
			Expect(err).Should(BeNil(), "Not able to load fixture form submission.")

			// insert the form into the coral database
			s, errSub = submission.Upsert(tests.Context, db, fixtureFormSubmissionInput)
		})

		Context("with appropiate context", func() {
			It("should not give an error", func() {
				Expect(errSub).Should(BeNil())
				Expect(s).ShouldNot(BeNil())
			})
			It("should return a submission for its form", func() {
				expectedFormID := "577c18f4a969c805f7f8c889"
				Expect(s.FormID.Hex()).Should(Equal(expectedFormID))
			})
		})
	})

})
