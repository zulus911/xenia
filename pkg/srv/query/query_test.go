package query_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/coralproject/shelf/pkg/srv/mongo"
	"github.com/coralproject/shelf/pkg/srv/query"
	"github.com/coralproject/shelf/pkg/tests"
)

var context = "testing"

func init() {
	tests.Init()
}

//==============================================================================

// removeSets is used to clear out all the test sets from the collection.
// All test query sets must start with QSTEST in their name.
func removeSets(ses *mgo.Session) error {
	f := func(c *mgo.Collection) error {
		q := bson.M{"name": bson.RegEx{Pattern: "QTEST"}}
		_, err := c.RemoveAll(q)
		return err
	}

	err := mongo.ExecuteDB(context, ses, "query_sets", f)
	if err != mgo.ErrNotFound {
		return err
	}

	return nil
}

// getFixture retrieves a query record from the filesystem.
func getFixture(filePath string) (*query.Set, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var qs query.Set
	err = json.NewDecoder(file).Decode(&qs)
	if err != nil {
		return nil, err
	}

	return &qs, nil
}

//==============================================================================

// TestCreateQuery tests if we can create a query record in the db.
func TestCreateQuery(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	ses := mongo.GetSession()
	defer ses.Close()

	defer func() {
		if err := removeSets(ses); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the query set : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the query set.", tests.Success)
	}()

	t.Log("Given the need to save a query set into the database.")
	{
		t.Log("\tWhen using fixture", fixture)
		{
			if err := query.CreateSet(context, ses, *qs1); err != nil {
				t.Fatalf("\t%s\tShould be able to create a query set : %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to create a query set.", tests.Success)
			}

			qs2, err := query.GetSetByName(context, ses, qs1.Name)
			if err != nil {
				t.Errorf("\t%s\tShould be able to retrieve the query set : %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to retrieve the query set.", tests.Success)
			}

			if qs1.Name != qs2.Name {
				t.Errorf("\t%s\tShould be able to get back the same query set.", tests.Failed)
				t.Logf("\t%+v", *qs1)
				t.Logf("\t%+v", *qs2)
			} else {
				t.Logf("\t%s\tShould be able to get back the same query set.", tests.Success)
			}

			if qs1.Enabled != qs2.Enabled {
				t.Errorf("\t%s\tShould have Enabled property set to true.", tests.Failed)
				t.Logf("\t%+v", *qs1)
				t.Logf("\t%+v", *qs2)
			} else {
				t.Logf("\t%s\tShould have Enabled property set to true.", tests.Success)
			}

			if qs1.Description == "" && qs2.Description == "" && qs1.Description != qs2.Description {
				t.Errorf("\t%s\tShould have a description for the query.", tests.Failed)
				t.Logf("\t%+v", *qs1)
				t.Logf("\t%+v", *qs2)
			} else {
				t.Logf("\t%s\tShould have a description for the query.", tests.Success)
			}

			if q1len, q2len := len(qs1.Params), len(qs2.Params); q1len != 1 && q2len != 1 && q1len != q2len {
				t.Errorf("\t%s\tShould have atleast one param in query params list.", tests.Failed)
				t.Logf("\t%+v", qs1.Params)
				t.Logf("\t%+v", qs2.Params)
			} else {
				t.Logf("\t%s\tShould have atleast one param in query params list.", tests.Success)

				for ind, param1 := range qs1.Params {
					param2 := qs2.Params[ind]
					if !reflect.DeepEqual(param1, param2) {
						t.Errorf("\t%s\tShould be able to validate query params.", tests.Failed)
						t.Logf("\t%+v", param1)
						t.Logf("\t%+v", param2)
					} else {
						t.Logf("\t%s\tShould be able to validate query params.", tests.Success)
					}
				}
			}

			if q1len, q2len := len(qs1.Queries), len(qs2.Queries); q1len != 2 && q2len != 2 && q1len != q2len {
				t.Errorf("\t%s\tShould have two query rule in queryset rule list.", tests.Failed)
				t.Logf("\t%+v", qs1.Params)
				t.Logf("\t%+v", qs2.Params)
			} else {
				t.Logf("\t%s\tShould have two query rule in queryset rule list.", tests.Success)

				for ind, qu := range qs1.Queries {
					qu2 := qs2.Queries[ind]
					if !reflect.DeepEqual(qu, qu2) {
						t.Errorf("\t%s\tShould be able to validate query rule in query list.", tests.Failed)
						t.Logf("\t%+v", qu)
						t.Logf("\t%+v", qu2)
					} else {
						t.Logf("\t%s\tShould be able to validate query rule in query list.", tests.Success)
					}
				}
			}
		}
	}
}

// TestGetSetNames validates retrieval of query.Set record names.
func TestGetSetNames(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	qsName := "spending_advice"

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	ses := mongo.GetSession()
	defer ses.Close()

	defer func() {
		if err := removeSets(ses); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the query set : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the query set.", tests.Success)
	}()

	t.Log("Given the need to retrieve a list of query sets.")
	{
		t.Log("\tWhen using fixture", fixture)
		{
			if err := query.CreateSet(context, ses, *qs); err != nil {
				t.Fatalf("\t%s\tShould be able to create a query set : %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to create a query set.", tests.Success)
			}

			qs.Name = qs.Name + "2"
			if err := query.CreateSet(context, ses, *qs); err != nil {
				t.Fatalf("\t%s\tShould be able to create a second query set : %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to create a second query set.", tests.Success)
			}

			names, err := query.GetSetNames(context, ses)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve the query set names : %v", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to retrieve the query set names", tests.Success)
			}

			if len(names) != 2 {
				t.Errorf("\t%s\tShould have two query sets : %s", tests.Failed, names)
			} else {
				t.Logf("\t%s\tShould have atleast one query record name: %s", tests.Success, names)
			}

			if !strings.Contains(names[0], qsName) || !strings.Contains(names[1], qsName) {
				t.Errorf("\t%s\tShould have \"%s\" in the name.", tests.Failed, qsName)
			} else {
				t.Logf("\t%s\tShould have \"%s\" in the name.", tests.Success, qsName)
			}
		}
	}
}

// TestUpdateSet set validates update operation of a given record.
func TestUpdateSet(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	ses := mongo.GetSession()
	defer ses.Close()

	defer func() {
		if err := removeSets(ses); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the query set : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the query set.", tests.Success)
	}()

	t.Log("Given the need to update a query set into the database.")
	{
		t.Log("\tWhen using fixture", fixture)
		{
			if err := query.CreateSet(context, ses, *qs1); err != nil {
				t.Errorf("\t%s\tShould be able to create a query set : %s", tests.Failed, err)
				t.Logf("\t%+v", *qs1)
			} else {
				t.Logf("\t%s\tShould be able to create a query set.", tests.Success)
			}

			qs2 := *qs1
			qs2.Params = append(qs2.Params, query.SetParam{
				Name:    "group",
				Default: "1",
				Desc:    "provides the group number for the query script",
			})

			if err := query.UpdateSet(context, ses, qs2); err != nil {
				t.Errorf("\t%s\tShould be able to update a query set record: %s", tests.Failed, err)
				t.Logf("\t%+v", qs2)
			} else {
				t.Logf("\t%s\tShould be able to update a query set record.", tests.Success)
			}

			updSet, err := query.GetSetByName(context, ses, qs2.Name)
			if err != nil {
				t.Errorf("\t%s\tShould be able to retrieve a query set record: %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to retrieve a query set record.", tests.Success)
			}

			if updSet.Name != qs2.Name && updSet.Name != qs1.Name {
				t.Errorf("\t%s\tShould be able to get back the same query set: %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould be able to get back the same query set", tests.Success)
			}

			if l1, l2, l3 := len(qs1.Params), len(qs2.Params), len(updSet.Params); l1 == l2 && (l3 < l1) {
				t.Errorf("\t%s\tShould have unequal one large param list in updated query set: %s", tests.Failed, err)
			} else {
				t.Logf("\t%s\tShould have unequal one large param list in updated query set.", tests.Success)

				param1 := qs1.Params[0]
				param2 := qs2.Params[0]
				uParam := updSet.Params[0]

				if !reflect.DeepEqual(uParam, param1) || !reflect.DeepEqual(uParam, param2) {
					t.Errorf("\t%s\tShould be abe to validate the query param at index 0: %s", tests.Failed, err)
					t.Logf("\t%+v", param1)
					t.Logf("\t%+v", param2)
					t.Logf("\t%+v", uParam)
				} else {
					t.Logf("\t%s\tShould be abe to validate the query param at index 0.", tests.Success)
				}
			}
		}
	}
}

// TestDeleteSet validates the removal of a query from the database.
func TestDeleteSet(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	qsName := "QTEST_spending_advice"

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	ses := mongo.GetSession()
	defer ses.Close()

	defer func() {
		if err := removeSets(ses); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the query set : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the query set.", tests.Success)
	}()

	t.Log("Given the need to delete a query set in the database.")
	{
		t.Log("\tWhen using fixture", fixture)
		{
			if err := query.CreateSet(context, ses, *qs1); err != nil {
				t.Errorf("\t%s\tShould be able to create a query set : %s", tests.Failed, err)
				t.Logf("\t%+v", *qs1)
			} else {
				t.Logf("\t%s\tShould be able to create a query set.", tests.Success)
			}

			if err := query.DeleteSet(context, ses, qsName); err != nil {
				t.Fatalf("\t%s\tShould be able to delete a query set using its name[%s]: %s", tests.Failed, qsName, err)
				t.Logf("\t%+v", fmt.Sprintf("{name: %s}", qs1.Name))
			} else {
				t.Logf("\t%s\tShould be able to delete a query set using its name[%s]:", tests.Success, qsName)
			}

			if _, err := query.GetSetByName(context, ses, qsName); err == nil {
				t.Fatalf("\t%s\tShould be able to validate query set with Name[%s] does not exists: %s", tests.Failed, qsName, errors.New("Record Exists"))
			} else {
				t.Logf("\t%s\tShould be able to validate query set with Name[%s] does not exists:", tests.Success, qsName)
			}

		}
	}
}

// TestUnknownName validates the behaviour of the query API when using a invalid/
// unknown query name.
func TestUnknownName(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	qsName := "QTEST_spending_desire"

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	ses := mongo.GetSession()
	defer ses.Close()

	defer func() {
		if err := removeSets(ses); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the query set : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the query set.", tests.Success)
	}()

	t.Log("Given the need to validate bad query name response.")
	{
		t.Log("\tWhen using fixture", fixture)
		{
			if err := query.CreateSet(context, ses, *qs1); err != nil {
				t.Fatalf("\t%s\tShould be able to create a query set : %s", tests.Failed, err)
				t.Logf("\t%+v", *qs1)
			} else {
				t.Logf("\t%s\tShould be able to create a query set.", tests.Success)
			}

			if _, err := query.GetSetByName(context, ses, qsName); err == nil {
				t.Fatalf("\t%s\tShould be able to validate query set with Name[%s] does not exists: %s", tests.Failed, qsName, errors.New("Record Exists"))
			} else {
				t.Logf("\t%s\tShould be able to validate query set with Name[%s] does not exists.", tests.Success, qsName)
			}

			if err := query.DeleteSet(context, ses, qsName); err == nil {
				t.Fatalf("\t%s\tShould be able to validate query set with Name[%s] can not be deleted: %s", tests.Failed, qsName, errors.New("Record Exists"))
			} else {
				t.Logf("\t%s\tShould be able to validate query set with Name[%s] can not be deleted.", tests.Success, qsName)
			}
		}

	}
}

// TestAPIFailure validates the failure of the api using a nil session.
func TestAPIFailure(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	qsName := "QTEST_spending_desire"

	const fixture = "./fixtures/spending_advice.json"
	qs1, err := getFixture(fixture)
	if err != nil {
		t.Fatalf("\t%s\tShould load query record from file : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould load query record from file.", tests.Success)

	t.Log("Given the need to to validate failure of API with bad session.")
	{
		t.Log("When giving a nil session")
		{
			if err := query.CreateSet(context, nil, *qs1); err == nil {
				t.Errorf("\t%s\tShould be refused create by api with bad session", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be refused create by api with bad session: %s", tests.Success, err)
			}

			if err := query.UpdateSet(context, nil, *qs1); err == nil {
				t.Errorf("\t%s\tShould be refused update by api with bad session", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be refused update by api with bad session: %s", tests.Success, err)
			}

			if _, err := query.GetSetByName(context, nil, qsName); err == nil {
				t.Errorf("\t%s\tShould be refused get request by api with bad session", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be refused get request by api with bad session: %s", tests.Success, err)
			}

			if _, err := query.GetSetNames(context, nil); err == nil {
				t.Errorf("\t%s\tShould be refused names request by api with bad session", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be refused names request by api with bad session: %s", tests.Success, err)
			}

			if err := query.DeleteSet(context, nil, qsName); err == nil {
				t.Errorf("\t%s\tShould be refused delete by api with bad session", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be refused delete by api with bad session: %s", tests.Success, err)
			}
		}
	}
}
