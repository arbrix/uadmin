package uadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// CSVImporterHandler handles CSV files
func CSVImporterHandler(w http.ResponseWriter, r *http.Request, session *Session) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	dataJSON := r.FormValue("data")
	if dataJSON == "" || dataJSON == "[]" {
		Trail(ERROR, "no csv data is provided")
        http.Error(w, "no csv data is provided", http.StatusBadRequest)
		return
	}

	// a csv file schema is the fields in the order they are defined in the model
	// a csv file should respect this order, the data should have two additional fields: id and language
	// the id should be the same for all the languages (translates) of the same model instance
	var csvFileRows []string
	if err := json.Unmarshal([]byte(dataJSON), &csvFileRows); err != nil {
		Trail(ERROR, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelDataMapping, err := getModelDataMapping(csvFileRows)
	if err != nil {
		Trail(ERROR, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelName := r.FormValue("m")
	s, _ := getSchema(modelName)
	fields := getFieldsList(s) // we rely on this order

	for _, objectDescription := range modelDataMapping {
		ok, err := objectExists(modelName, objectDescription, fields)
		if err != nil {
			Trail(ERROR, err.Error())
			http.Error(w, "failed to process model data", http.StatusInternalServerError)
		}
		if ok {
			continue
		}

		model, err := getPopulatedModel(modelName, objectDescription, fields)
		if err != nil {
			Trail(ERROR, "failed to process model %v", err.Error())
			http.Error(w, "failed to process model", http.StatusBadRequest)
			return
		}
		SaveRecord(model)
	}

	response := map[string]string{
		"message": "CSV data successfully imported",
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "failed to create response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// language-values mapping for the csv file object description
type csvEntry struct {
	Langs  []string
	Fields map[string][]string
}

// returns a list of models descriptions from the provided csv file data
// the incoming csv data should be in the following format:
//    `idx;lang;the;rest;fields`
//    idx — index of the object entry in the csv file, the same for all the languages
//    lang — particular language for the object entry in the csv file
func getModelDataMapping(csvFileRows []string) ([]csvEntry, error) {
	if len(csvFileRows) == 0 {
		return nil, fmt.Errorf("no csv file rows")
	}

	csvEntries := map[string]csvEntry{}
	ids := []string{}
	for _, row := range csvFileRows {
		rowData := strings.Split(row, ";")
		if len(rowData) < 3 { // expected at least 3 fields: row id, lang, model field (one or more)
			return nil, fmt.Errorf("csv file row doesn't have any model data")
		}

		// collect all the languages and data for the same object
		rowID := rowData[0]
		rowLang := rowData[1]
		if entry, ok := csvEntries[rowID]; ok {
			// add another one language
			entry.Langs = append(entry.Langs, rowLang)
			entry.Fields[rowLang] = rowData[2:]
			csvEntries[rowID] = entry
		} else {
			ids = append(ids, rowID)
			// add a new entry
			csvEntries[rowID] = csvEntry{
				Langs:  []string{rowLang},
				Fields: map[string][]string{rowLang: rowData[2:]},
			}
		}
	}

	data := []csvEntry{}
	for _, id := range ids {
		data = append(data, csvEntries[id])
	}

	return data, nil
}

type fieldDescriptor struct {
	Name        string
	Type        string
	FK          bool
	FKModelName string
}

// returns a list of a model fields in the order they are defined, marks foreign keys in the list
func getFieldsList(s ModelSchema) []fieldDescriptor {
	var list []fieldDescriptor
	for _, f := range s.Fields {
		if f.Name == "ID" { // skip the base model field
			continue
		}
		fd := fieldDescriptor{
			Name: f.Name,
			Type: f.Type,
		}
		if f.Type == "fk" {
			//TODO: this is a struct! How to find out what a struct's field is used in a csv file?..
			// f.Name here is the name for this FK in the parent model, it's not the same as f.TypeName (FK struct's name,
			// lower case of which is the fk's model name)
			fd.FK = true
			fd.FKModelName = strings.ToLower(f.TypeName)
		}
		list = append(list, fd)
	}
	return list
}

func objectExists(modelName string, objectDescription csvEntry, fieldsList []fieldDescriptor) (bool, error) {
	lang := objectDescription.Langs[0]
	fields := objectDescription.Fields[lang]

	var conditions []string
	var values []interface{}
	for idx, fieldDesc := range fieldsList {
		if fieldDesc.Type == "fk" {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s::jsonb->>? = ?", toSnakeCase(fieldDesc.Name)))
		values = append(values, lang, fields[idx])
	}

	var model reflect.Value
	model, ok := NewModel(modelName, true)
	if !ok {
		return false, fmt.Errorf("bad model: %s", modelName)
	}

	query := strings.Join(conditions, " AND ")
	err := Get(model.Interface(), query, values...)
	if err != nil && err.Error() != "record not found" {
		Trail(ERROR, "query '%s' is failed: %v", query, err)
		return false, err
	}
	if err == nil && GetID(model) != 0 {
		return true, nil
	}

	return false, nil
}

func getPopulatedModel(modelName string, objectDescription csvEntry, fieldsList []fieldDescriptor) (reflect.Value, error) {
	nilValue := reflect.ValueOf(nil)
	model, ok := NewModel(modelName, true)
	if !ok {
		return nilValue, fmt.Errorf("bad model: %s", modelName)
	}

	for idx, fieldDesc := range fieldsList {
		if field := model.Elem().FieldByName(fieldDesc.Name); field.IsValid() && field.CanSet() {
			langToFieldsMap := map[string]string{} // will be marshaled to a string like `{"en":"value"}`
			for _, lang := range objectDescription.Langs {
				fields := objectDescription.Fields[lang] // the values for all the fields of this model description in this lang
				langToFieldsMap[lang] = fields[idx]
			}

			fieldsMultilangValueJSON, err := json.Marshal(langToFieldsMap)
			if err != nil {
				return nilValue, err
			}

			if fieldDesc.Type == "fk" {
				m, ok := NewModel(fieldDesc.FKModelName, true)
				if !ok {
					return nilValue, fmt.Errorf("can't get %s model", fieldDesc.FKModelName)
				}

				// TODO: this works for one use-case only. To make it general, we need a way to
				// pass FK FieldName with this value (see `data[idx]` above) in a csv file
				hardcoded := "name"
				q := fmt.Sprintf("%s::jsonb->>? = ?", toSnakeCase(hardcoded))
				fields := langToFieldsMap[objectDescription.Langs[0]]
				err := Get(m.Interface(), q, objectDescription.Langs[0], fields)
				if err != nil && err.Error() != "record not found" {
					Trail(ERROR, "query '%s' is failed: %v", q, err)
					return nilValue, err
				}
				if (err != nil && err.Error() != "record not found") || GetID(m) == 0 {
					// TODO: probably, we want to avoid creating FK object in this handler since such a struct might require data we don't have here
					// TODO: this works for one use-case only.
					Trail(INFO, "no record for: '%s', going to create a new one", fields)
					hardcodedFN := "Name"
					// foreign key model's field
					if field := m.Elem().FieldByName(hardcodedFN); field.IsValid() && field.CanSet() {
						field.SetString(string(fieldsMultilangValueJSON))
					}
					SaveRecord(m)
				}

				field.Set(m.Elem())
				continue
			}

			field.SetString(string(fieldsMultilangValueJSON))
		}
	}

	return model, nil
}
