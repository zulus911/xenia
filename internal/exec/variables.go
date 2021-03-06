package exec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ardanlabs/kit/log"

	"gopkg.in/mgo.v2/bson"
)

// ProcessVariables walks the document performing variable substitutions.
// This function is exported because it is accessed by the tstdata package.
func ProcessVariables(context interface{}, commands map[string]interface{}, vars map[string]string, results map[string]interface{}) error {

	// commands: Contains the mongodb pipeline with any extenstions.
	// vars    : Key/Value pairs passed into the set execution for variable substituion.
	// results : Any result from previous sets that have been saved.

	// A map of keys that may need to be replaced.
	keyReplace := make(map[string]string)

	for key, value := range commands {

		// Does the key have a variable syntax.
		idx := strings.IndexByte(key, '{')
		if idx != -1 {
			if err := fldSub(context, key, vars, keyReplace); err != nil {
				return err
			}
		}

		// Test for the type of value we have.
		switch doc := value.(type) {

		// We have another document.
		case map[string]interface{}:
			if err := ProcessVariables(context, doc, vars, results); err != nil {
				return err
			}

		// We have a string value so check it.
		case string:
			if doc != "" && doc[0] == '#' {
				if err := valSub(context, key, doc, commands, vars, results); err != nil {
					return err
				}
			}

		// We have an array of values.
		case []interface{}:

			// Iterate over the array of values.
			for _, subDoc := range doc {

				// What type of subDoc is this array made of.
				switch arrDoc := subDoc.(type) {

				// We have another document.
				case map[string]interface{}:
					if err := ProcessVariables(context, arrDoc, vars, results); err != nil {
						return err
					}

				// We have a string value so check it.
				case string:
					if arrDoc != "" && arrDoc[0] == '#' {
						if err := valSub(context, key, arrDoc, commands, vars, results); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// Are there fields to replace.
	if len(keyReplace) > 0 {
		for k, v := range keyReplace {

			// Add a new key with this data.
			commands[v] = commands[k]

			// Remove the old key.
			delete(commands, k)
		}
	}

	return nil
}

// fldSub appends to the replace map the fields that need to change and what the
// new field name is.
func fldSub(context interface{}, key string, vars map[string]string, replace map[string]string) error {

	// Before: statstics.comments.{dimension}.{commentStatus}.{value}
	// After:  statstics.comments.dim.cstat.v

	parts := strings.Split(key, ".")
	for i, p := range parts {

		// If there is not variable, move to the next part.
		if p[0] != '{' {
			continue
		}

		l := len(p) - 1

		if l < 3 {
			err := fmt.Errorf("Invalid field variable : %q", p)
			log.Error(context, "fldSub", err, "validating field variable length")
			return err
		}

		if p[l] != '}' {
			err := fmt.Errorf("Invalid field variable : %q", p)
			log.Error(context, "fldSub", err, "validating field variable syntax")
			return err
		}

		fld := p[1:l]

		nFld, exists := vars[fld]
		if !exists {
			err := fmt.Errorf("Field variable does not exist : %q", fld)
			log.Error(context, "fldSub", err, "field variable lookup")
			return err
		}

		parts[i] = nFld
	}

	replace[key] = strings.Join(parts, ".")
	return nil
}

// valSub replaces variables inside the command set with values.
func valSub(context interface{}, key, variable string, commands map[string]interface{}, vars map[string]string, results map[string]interface{}) error {

	// Before: {"field": "#number:variable_name"}  After: {"field": 1234}
	// key: "field"  variable:"#cmd:variable_name"

	// Remove the # characters from the left.
	value := variable[1:]

	// Find the first instance of the separator.
	idx := strings.IndexByte(value, ':')
	if idx == -1 {
		err := fmt.Errorf("Invalid variable format %q, missing :", variable)
		log.Error(context, "varSub", err, "Parsing variable")
		return err
	}

	// Split the key and variable apart.
	cmd := value[0:idx]
	vari := value[idx+1:]

	switch key {
	case "$in":
		if len(cmd) != 6 || cmd[0:4] != "data" {
			err := fmt.Errorf("Invalid $in command %q, missing \"data\" keyword or malformed", cmd)
			log.Error(context, "varSub", err, "$in command processing")
			return err
		}

		v, err := dataLookup(context, cmd[5:6], vari, results)
		if err != nil {
			return err
		}

		commands[key] = v
		return nil

	default:
		v, err := varLookup(context, cmd, vari, vars, results)
		if err != nil {
			return err
		}

		commands[key] = v
		return nil
	}
}

// varLookup looks up variables and returns their values as the specified type.
func varLookup(context interface{}, cmd, variable string, vars map[string]string, results map[string]interface{}) (interface{}, error) {

	// {"field": "#cmd:variable"}
	// Before: {"field": "#number:variable_name"}  		After: {"field": 1234}
	// Before: {"field": "#string:variable_name"}  		After: {"field": "value"}
	// Before: {"field": "#date:variable_name"}    		After: {"field": time.Time}
	// Before: {"field": "#objid:variable_name"}   		After: {"field": mgo.ObjectId}
	// Before: {"field": "#regex:/pattern/<options>"}   After: {"field": bson.RegEx}
	// Before: {"field": "#since:3600"}   				After: {"field": time.Time}
	// Before: {"field": "#data.0:doc.station_id"}   	After: {"field": "23453"}

	// If the variable does not exist, use the variable straight up.
	param, exists := vars[variable]
	if !exists {
		param = variable
	}

	// Do we have a command that is not known.
	if len(cmd) < 4 {
		err := fmt.Errorf("Unknown command %q", cmd)
		log.Error(context, "varLookup", err, "Checking cmd is proper length")
		return nil, err
	}

	// Let's perform the right action per command.
	switch cmd[0:4] {
	case "numb":
		return number(context, param)

	case "stri":
		return param, nil

	case "date":
		return isoDate(context, param)

	case "obji":
		return objID(context, param)

	case "rege":
		return regExp(context, param)

	case "time":
		return adjTime(context, param)

	case "data":
		if len(cmd) == 6 {
			return dataLookup(context, cmd[5:6], param, results)
		}

		err := errors.New("Data command is missing the operator")
		log.Error(context, "varLookup", err, "Checking cmd is data")
		return nil, err

	default:
		err := fmt.Errorf("Unknown command %q", cmd)
		log.Error(context, "varLookup", err, "Checking cmd in default case")
		return nil, err
	}
}

// dataLookup looks up data from the saved results based on the data operation
// and the lookup value.
func dataLookup(context interface{}, dataOp, lookup string, results map[string]interface{}) (interface{}, error) {

	// We you need an array to be substitued.						// We you need a single value to be substitued, select an index.
	// Before: {"field" : {"$in": "#data.*:list.station_id"}}}		// Before: {"field" : "#data.0:list.station_id"}
	// After : {"field" : {"$in": ["42021"]}}						// After : {"field" : "42021"}
	//      dataOp : "*"                                         	//      dataOp : 0
	//  	lookup : "list.station_id"							    //  	lookup : "list.station_id"
	//  	results: {"list": [{"station_id":"42021"}]}				//  	results: {"list": [{"station_id":"42021"}, {"station_id":"23567"}]}

	// Find the result data based on the lookup and the field lookup.
	data, field, err := findResultData(context, lookup, results)
	if err != nil {
		return "", err
	}

	// How many documents do we have.
	l := len(data)

	// If there is no data for the lookup.
	if l == 0 {
		err := errors.New("The results contain no documents")
		log.Error(context, "dataLookup", err, "Checking length")
		return "", err
	}

	// Do we need to return an array.
	if dataOp == "*" {

		// We need to create an array of the values.
		var array []interface{}
		for _, doc := range data {

			// Find the value for the specified field.
			fldValue, err := docFieldLookup(context, doc, field)
			if err != nil {
				return "", err
			}

			// Append the value to the array.
			array = append(array, fldValue)
		}

		return array, nil
	}

	// Convert the index position to an int.
	index, err := strconv.Atoi(dataOp)
	if err != nil {
		err = fmt.Errorf("Invalid operator command operator %q", dataOp)
		log.Error(context, "dataLookup", err, "Index conversion")
		return "", err
	}

	// We can't ask for a position we don't have.
	if index > l-1 {
		err := fmt.Errorf("Index \"%d\" out of range, total \"%d\"", index, l)
		log.Error(context, "dataLookup", err, "Index range check")
		return "", err
	}

	// Find the value for the specified field.
	fldValue, err := docFieldLookup(context, data[index], field)
	if err != nil {
		return "", err
	}

	return fldValue, nil
}

// findResultData process the lookup against the results. Returns the result if
// found and the field name for location the field from the results later.
func findResultData(context interface{}, lookup string, results map[string]interface{}) ([]bson.M, string, error) {

	// lookup: "station.station_id"		lookup: "list.condition.wind_string"
	// 		key  :   station				key  : list
	//		field: station_id				field: condition.wind_string

	// Split the lookup into the data key and document field key.
	idx := strings.IndexByte(lookup, '.')
	if idx == -1 {
		err := fmt.Errorf("Invalid formated lookup %q", lookup)
		log.Error(context, "findResultData", err, "Parsing lookup")
		return nil, "", err
	}

	// Extract the key and field.
	key := lookup[0:idx]
	field := lookup[idx+1:]

	// Find the result the user is looking for.
	data, exists := results[key]
	if !exists {
		err := fmt.Errorf("Key %q not found in saved results", key)
		log.Error(context, "findResultData", err, "Finding results")
		return nil, "", err
	}

	// Extract the concrete type from the interface.
	values, ok := data.([]bson.M)
	if !ok {
		err := errors.New("** FATAL : Expected the result to be an array of documents")
		log.Error(context, "findResultData", err, "Type assert results : %T", data)
		return nil, "", err
	}

	return values, field, nil
}

// docFieldLookup recurses the document for the specified field and returns its value.
func docFieldLookup(context interface{}, doc map[string]interface{}, field string) (interface{}, error) {

	// condition.location.type

	// Extract the first field for lookup and the remaining fields
	// if we need to recurse deeper into the result document.
	var nextFld string
	idx := strings.IndexByte(field, '.')
	if idx > 0 {
		nextFld = field[idx+1:]
		field = field[0:idx]
	}

	// Look up the first field.
	fldValue, exists := doc[field]
	if !exists {
		err := fmt.Errorf("Field %q not found", field)
		log.Error(context, "docFieldLookup", err, "Document field lookup")
		return nil, err
	}

	// When we have found the last field, we have the data.
	if idx == -1 {
		return fldValue, nil
	}

	// We need to recuse the result for the next field.

	// The field data we found needs to be a document.
	fldDoc, ok := fldValue.(bson.M)
	if !ok {
		err := fmt.Errorf("Field value is a %T and not a bson document", fldValue)
		log.Error(context, "docFieldLookup", err, "Type assert for document")
		return nil, err
	}

	// Find the next field and its value.
	fldValue, err := docFieldLookup(context, fldDoc, nextFld)
	if err != nil {
		return nil, err
	}

	return fldValue, nil
}

//==============================================================================

// number is ahelper function to convert the value to an integer.
func number(context interface{}, value string) (int, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		err = fmt.Errorf("Parameter %q is not a number", value)
		log.Error(context, "varLookup", err, "Index conversion")
		return 0, err
	}
	return i, nil
}

// isoDate is a helper function to convert the internal extension for dates
// into a BSON date. We convert the following string
func isoDate(context interface{}, value string) (time.Time, error) {
	var parse string

	switch len(value) {
	case 10:
		parse = "2006-01-02"
	case 24:
		parse = "2006-01-02T15:04:05.999Z"
	case 23:
		parse = "2006-01-02T15:04:05.999"
	default:
		err := fmt.Errorf("Invalid date value %q", value)
		log.Error(context, "isoDate", err, "Selecting date parse string")
		return time.Time{}, err
	}

	dateTime, err := time.Parse(parse, value)
	if err != nil {
		log.Error(context, "isoDate", err, "Parsing date string")
		return time.Time{}, err
	}

	return dateTime, nil
}

// objID is a helper function to convert a string that represents a Mongo
// Object Id into a bson ObjectId type.
func objID(context interface{}, value string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(value) {
		err := fmt.Errorf("Objectid %q is invalid", value)
		log.Error(context, "objID", err, "Checking obj validity")
		return bson.ObjectId(""), err
	}

	return bson.ObjectIdHex(value), nil
}

// regExp is a helper function to process regex commands.
func regExp(context interface{}, value string) (bson.RegEx, error) {
	idx := strings.IndexByte(value[1:], '/')
	if value[0] != '/' || idx == -1 {
		err := fmt.Errorf("Parameter %q is not a regular expression", value)
		log.Error(context, "varLookup", err, "Regex parsing")
		return bson.RegEx{}, err
	}

	pattern := value[1 : idx+1]
	l := len(pattern) + 2

	var options string
	if l < len(value) {
		options = value[l:]
	}

	return bson.RegEx{Pattern: pattern, Options: options}, nil
}

// adjTime is a helper function to take the current time and adjust it based
// on the provided value.
func adjTime(context interface{}, value string) (time.Time, error) {

	// The default value is in seconds unless overridden.
	// #time:0      Current date/time
	// #time:-3600  3600 seconds in the past
	// #time:3m		3 minutes in the future.

	// Possible duration types.
	// "ns": int64(Nanosecond),
	// "us": int64(Microsecond),
	// "ms": int64(Millisecond),
	// "s":  int64(Second),
	// "m":  int64(Minute),
	// "h":  int64(Hour),

	// Do we have a single value?
	if len(value) == 1 {
		val, err := strconv.Atoi(value[0:1])
		if err != nil {
			return time.Time{}.UTC(), fmt.Errorf("Invalid duration : %q", value[0:1])
		}

		if val == 0 {
			return time.Now().UTC(), nil
		}

		return time.Now().Add(time.Duration(val) * time.Second).UTC(), nil
	}

	// Do we have a duration type and where does the
	// actual duration value end
	var typ string
	var end int

	// The end byte position for the last character in the string.
	ePos := len(value) - 1

	// Look at the very last character.
	t := value[ePos:]
	switch t {

	// Is this a minute or hour? [3m]
	case "m", "h":
		typ = t
		end = ePos // Position of last chr in value.

	// Is this a second or other duration? [3s or 3us]
	case "s":
		typ = t    // s for 3s
		end = ePos // 3 for 3s

		// Is this smaller than a second? [ns, us, ms]
		if len(value) > 2 {
			t := value[ePos-1 : ePos]
			switch t {
			case "n", "u", "m":
				typ = value[ePos-1:] // us for 3us
				end = ePos - 1       // 3 for 3us
			}
		}

	default:
		typ = "s"      // s for 3600
		end = ePos + 1 // 0 for 3600
	}

	// Check if we are to negative the value.
	var start int
	if value[0] == '-' {
		start = 1
	}

	// Check the remaining bytes is an integer value.
	val, err := strconv.Atoi(value[start:end])
	if err != nil {
		return time.Time{}.UTC(), fmt.Errorf("Invalid duration : %q", value[start:end])
	}

	// Do we have to negate the value?
	if start == 1 {
		val *= -1
	}

	// Calcuate the time value.
	switch typ {
	case "ns":
		return time.Now().Add(time.Duration(val) * time.Nanosecond).UTC(), nil
	case "us":
		return time.Now().Add(time.Duration(val) * time.Microsecond).UTC(), nil
	case "ms":
		return time.Now().Add(time.Duration(val) * time.Millisecond).UTC(), nil
	case "m":
		return time.Now().Add(time.Duration(val) * time.Minute).UTC(), nil
	case "h":
		return time.Now().Add(time.Duration(val) * time.Hour).UTC(), nil
	default:
		return time.Now().Add(time.Duration(val) * time.Second).UTC(), nil
	}
}
