package form_test

import (
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/tests"

	"github.com/coralproject/xenia/internal/form"
	"github.com/coralproject/xenia/internal/form/ffix"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {

	var (
		db *db.DB
	)

	BeforeEach(func() {
		// get a database from the fixtures
		db = ffix.SetTestDatabase()
	})

	AfterEach(func() {
		// empty database
		ffix.TearTestDatabase(db)
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

			if fixtureForm, err = ffix.GetFixtureForm(); err != nil {
				Expect(err).Should(BeNil(), "Not able to load fixture form.")
			}

			result, err = form.Upsert(tests.Context, db, fixtureForm)
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
		db = ffix.SetTestDatabase()
	})

	AfterEach(func() {
		// empty database
		ffix.TearTestDatabase(db)
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
			if fixtureForm, err = ffix.GetFixtureForm(); err != nil {
				Expect(err).Should(BeNil(), "Not able to load fixture form.")
			}

			// insert the form into the coral database
			if _, err = form.Upsert(tests.Context, db, fixtureForm); err != nil {
				Expect(err).Should(BeNil(), "Not able to create fixture form.")
			}

			fixtureForm.Status = "new"
			result, err = form.Upsert(tests.Context, db, fixtureForm)
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

})
