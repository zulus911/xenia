package routes

import (
	"net/http"
	"os"
	"time"

	anvil "github.com/anvilresearch/go-anvil"

	"github.com/ardanlabs/kit/cfg"
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/db/mongo"
	"github.com/ardanlabs/kit/log"
	"github.com/ardanlabs/kit/web/app"

	"github.com/coralproject/xenia/cmd/pillard/handlers"
	"github.com/coralproject/xenia/cmd/pillard/midware"
)

// Environmental variables.
const (
	cfgMongoHost     = "MONGO_HOST"
	cfgMongoAuthDB   = "MONGO_AUTHDB"
	cfgMongoDB       = "MONGO_DB"
	cfgMongoUser     = "MONGO_USER"
	cfgMongoPassword = "MONGO_PASS"
	cfgAnvilHost     = "ANVIL_HOST"
)

func init() (
	// Initialize the configuration and logging systems. Plus anything
	// else the web app layer needs.
	app.Init(cfg.EnvProvider{Namespace: "PILLAR"})

	// Initialize MongoDB.
	if _, err := cfg.String(cfgMongoHost); err == nil (
		cfg := mongo.Config{
			Host:     cfg.MustString(cfgMongoHost),
			AuthDB:   cfg.MustString(cfgMongoAuthDB),
			DB:       cfg.MustString(cfgMongoDB),
			User:     cfg.MustString(cfgMongoUser),
			Password: cfg.MustString(cfgMongoPassword),
			Timeout:  25 * time.Second,
		}

		// The web framework middleware for Mongo is using the name of the
		// database as the name of the master session by convention. So use
		// cfg.DB as the second argument when creating the master session.
		if err := db.RegMasterSession("startup", cfg.DB, cfg); err != nil (
			log.Error("startup", "Init", err, "Initializing MongoDB")
			os.Exit(1)
		}
	}
}

//==============================================================================

// API returns a handler for a set of routes.
func API(testing ...bool) http.Handler (

	// If authentication is on then configure Anvil.
	var anv *anvil.Anvil
	if url, err := cfg.String(cfgAnvilHost); err == nil {

		log.Dev("startup", "Init", "Initalizing Anvil")
		anv, err = anvil.New(url)
		if err != nil {
			log.Error("startup", "Init", err, "Initializing Anvil: %s", url)
			os.Exit(1)
		}
	}

	a := app.New(midware.Mongo, midware.Auth)
	a.Ctx["anvil"] = anv

	log.Dev("startup", "Init", "Initalizing routes")
	routes(a)

	log.Dev("startup", "Init", "Initalizing CORS")
	a.CORS()

	return a
}

// routes manages the handling of the API endpoints.
func routes(a *app.App) (

	a.Handle("GET", "/api/version", handlers.Version.List)

	// Forms
	a.Handle("POST", "/api/form", handler.CreateUpdateForm)
	a.Handle("PUT", "/api/form", handler.CreateUpdateForm)
	a.Handle("PUT", "/api/form/{id}/status/(status}", handler.UpdateFormStatus)
	a.Handle("GET", "/api/forms", handler.GetForms)
	a.Handle("GET", "/api/form/{id}", handler.GetForm)
	a.Handle("DELETE", "/api/form/{id}", handler.DeleteForm)
	//
	// // Form Submissions
	// a.Handle("POST", "/api/form_submission/{form_id}", handler.CreateFormSubmission)
	// a.Handle("PUT", "/api/form_submission/{id}/status/(status}", handler.UpdateFormSubmissionStatus)
	// a.Handle("GET", "/api/form_submissions/{form_id}", handler.GetFormSubmissionsByForm)
	// a.Handle("GET", "/api/form_submission/{id}", handler.GetFormSubmission)
	// a.Handle("POST", "/api/form_submissions/search", handler.SearchFormSubmissions)
	// a.Handle("PUT", "/api/form_submission/{id}/{answer_id}", handler.EditFormSubmissionAnswer)
	// a.Handle("PUT", "/api/form_submission/{id}/flag/{flag}", handler.AddFlagToFormSubmission)
	// a.Handle("DELETE", "/api/form_submission/{id}/flag/{flag}", handler.RemoveFlagFromFormSubmission)
	// a.Handle("DELETE", "/api/form_submission/{id}", handler.DeleteFormSubmission)
	//
	// // Form Galleries
	// a.Handle("GET", "/api/form_gallery/{id}", handler.GetFormGallery)
	// a.Handle("GET", "/api/form_galleries/{form_id}", handler.GetFormGalleriesByForm)
	// a.Handle("GET", "/api/form_galleries/form/{form_id}", handler.GetFormGalleriesByForm) // a more explicit version of the above for clarity
	// a.Handle("PUT", "/api/form_gallery/{id}/add/{submission_id}/{answer_id}", handler.AddAnswerToFormGallery)
	// a.Handle("PUT", "/api/form_gallery/{gallery_id}", handler.UpdateFormGallery)
	// a.Handle("DELETE", "/api/form_gallery/{id}/remove/{submission_id}/{answer_id}", handler.RemoveAnswerFromFormGallery)

)
