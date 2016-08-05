// Package handlers contains the handler logic for processing requests.
package handlers

/*

	This file contains handlers for web endpoint
	invocations of Ask services:

	* Forms
	* Form Submissions
	* Form Galleries

*/

// TO DO: ON BEHAVIOR PACKAGE, we neeed to migrate that package from Pillar to Xenia

import (
	"encoding/json"
	"net/http"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/web/app"
	"github.com/coralproject/xenia/internal/form"
)

// formHandle maintains the set of handlers for the form api.
type formHandle struct{}

// Form fronts the access to the form service functionality.
var Form formHandle

//==============================================================================

func unmarshalForm(c *app.Context) (*form.Form, error) {
	var f *form.Form
	err := json.NewDecoder(c.Request.Body).Decode(&f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (formHandle) CreateUpdateForm(c *app.Context) error {

	f, err := unmarshalForm(c)
	if err != nil {
		return err
	}

	rf, err := form.Upsert(c.SessionID, c.Ctx["DB"].(*db.DB), f)
	if err != nil {
		return err
	}

	c.Respond(rf, http.StatusOK)
	return nil
}

//
// func UpdateFormStatus(c *web.AppContext) {
// 	dbObject, err := service.UpdateFormStatus(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetForms(c *web.AppContext) {
// 	dbObject, err := service.GetForms(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetForm(c *web.AppContext) {
// 	dbObject, err := service.GetForm(c)
// 	doRespond(c, dbObject, err)
// }
//
// func DeleteForm(c *web.AppContext) {
// 	err := service.DeleteForm(c)
// 	doRespond(c, nil, err)
// }
//
// func CreateFormSubmission(c *web.AppContext) {
// 	dbObject, err := service.CreateFormSubmission(c)
// 	doRespond(c, dbObject, err)
// }
//
// func UpdateFormSubmissionStatus(c *web.AppContext) {
// 	dbObject, err := service.UpdateFormSubmissionStatus(c)
// 	doRespond(c, dbObject, err)
// }
//
// func EditFormSubmissionAnswer(c *web.AppContext) {
// 	dbObject, err := service.EditFormSubmissionAnswer(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetFormSubmissionsByForm(c *web.AppContext) {
// 	dbObject, err := service.GetFormSubmissionsByForm(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetFormSubmission(c *web.AppContext) {
// 	dbObject, err := service.GetFormSubmission(c)
// 	doRespond(c, dbObject, err)
// }
//
// func DeleteFormSubmission(c *web.AppContext) {
// 	err := service.DeleteFormSubmission(c)
// 	doRespond(c, nil, err)
// }
//
// func AddFlagToFormSubmission(c *web.AppContext) {
// 	dbObject, err := service.AddFlagToFormSubmission(c)
// 	doRespond(c, dbObject, err)
// }
//
// func RemoveFlagFromFormSubmission(c *web.AppContext) {
// 	dbObject, err := service.RemoveFlagFromFormSubmission(c)
// 	doRespond(c, dbObject, err)
// }
//
// func AddAnswerToFormGallery(c *web.AppContext) {
// 	dbObject, err := service.AddAnswerToFormGallery(c)
// 	doRespond(c, dbObject, err)
// }
//
// func RemoveAnswerFromFormGallery(c *web.AppContext) {
// 	dbObject, err := service.RemoveAnswerFromFormGallery(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetFormGalleriesByForm(c *web.AppContext) {
// 	dbObject, err := service.GetFormGalleriesByForm(c)
// 	doRespond(c, dbObject, err)
// }
//
// func GetFormGallery(c *web.AppContext) {
// 	dbObject, err := service.GetFormGallery(c)
// 	doRespond(c, dbObject, err)
// }
//
// func UpdateFormGallery(c *web.AppContext) {
// 	dbObject, err := service.UpdateFormGallery(c)
// 	doRespond(c, dbObject, err)
// }
//
// func SearchFormSubmissions(c *web.AppContext) {
//
// 	//Get the search string from request
// 	var search map[string]string
// 	json.NewDecoder(c.Body).Decode(&search)
// 	c.SetValue("search", search["search"])
//
// 	dbObject, err := service.SearchFormSubmissions(c)
//
// 	doRespond(c, dbObject, err)
// }
