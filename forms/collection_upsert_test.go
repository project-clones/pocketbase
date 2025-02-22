package forms_test

import (
	"encoding/json"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/spf13/cast"
)

func TestNewCollectionUpsert(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection := &models.Collection{}
	collection.Name = "test"
	collection.System = true
	listRule := "testview"
	collection.ListRule = &listRule
	viewRule := "test_view"
	collection.ViewRule = &viewRule
	createRule := "test_create"
	collection.CreateRule = &createRule
	updateRule := "test_update"
	collection.UpdateRule = &updateRule
	deleteRule := "test_delete"
	collection.DeleteRule = &deleteRule
	collection.Schema = schema.NewSchema(&schema.SchemaField{
		Name: "test",
		Type: schema.FieldTypeText,
	})

	form := forms.NewCollectionUpsert(app, collection)

	if form.Name != collection.Name {
		t.Errorf("Expected Name %q, got %q", collection.Name, form.Name)
	}

	if form.System != collection.System {
		t.Errorf("Expected System %v, got %v", collection.System, form.System)
	}

	if form.ListRule != collection.ListRule {
		t.Errorf("Expected ListRule %v, got %v", collection.ListRule, form.ListRule)
	}

	if form.ViewRule != collection.ViewRule {
		t.Errorf("Expected ViewRule %v, got %v", collection.ViewRule, form.ViewRule)
	}

	if form.CreateRule != collection.CreateRule {
		t.Errorf("Expected CreateRule %v, got %v", collection.CreateRule, form.CreateRule)
	}

	if form.UpdateRule != collection.UpdateRule {
		t.Errorf("Expected UpdateRule %v, got %v", collection.UpdateRule, form.UpdateRule)
	}

	if form.DeleteRule != collection.DeleteRule {
		t.Errorf("Expected DeleteRule %v, got %v", collection.DeleteRule, form.DeleteRule)
	}

	// store previous state and modify the collection schema to verify
	// that the form.Schema is a deep clone
	loadedSchema, _ := collection.Schema.MarshalJSON()
	collection.Schema.AddField(&schema.SchemaField{
		Name: "new_field",
		Type: schema.FieldTypeBool,
	})

	formSchema, _ := form.Schema.MarshalJSON()

	if string(formSchema) != string(loadedSchema) {
		t.Errorf("Expected Schema %v, got %v", string(loadedSchema), string(formSchema))
	}
}

func TestCollectionUpsertValidate(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		jsonData       string
		expectedErrors []string
	}{
		{"{}", []string{"name", "schema"}},
		{
			`{
				"name": "test ?!@#$",
				"system": true,
				"schema": [
					{"name":"","type":"text"}
				],
				"listRule": "missing = '123'",
				"viewRule": "missing = '123'",
				"createRule": "missing = '123'",
				"updateRule": "missing = '123'",
				"deleteRule": "missing = '123'"
			}`,
			[]string{"name", "schema", "listRule", "viewRule", "createRule", "updateRule", "deleteRule"},
		},
		{
			`{
				"name": "test",
				"system": true,
				"schema": [
					{"name":"test","type":"text"}
				],
				"listRule": "test='123'",
				"viewRule": "test='123'",
				"createRule": "test='123'",
				"updateRule": "test='123'",
				"deleteRule": "test='123'"
			}`,
			[]string{},
		},
	}

	for i, s := range scenarios {
		form := forms.NewCollectionUpsert(app, &models.Collection{})

		// load data
		loadErr := json.Unmarshal([]byte(s.jsonData), form)
		if loadErr != nil {
			t.Errorf("(%d) Failed to load form data: %v", i, loadErr)
			continue
		}

		// parse errors
		result := form.Validate()
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("(%d) Failed to parse errors %v", i, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("(%d) Expected error keys %v, got %v", i, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("(%d) Missing expected error key %q in %v", i, k, errs)
			}
		}
	}
}

func TestCollectionUpsertSubmit(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		existingName   string
		jsonData       string
		expectedErrors []string
	}{
		// empty create
		{"", "{}", []string{"name", "schema"}},
		// empty update
		{"demo", "{}", []string{}},
		// create failure
		{
			"",
			`{
				"name": "test ?!@#$",
				"system": true,
				"schema": [
					{"name":"","type":"text"}
				],
				"listRule": "missing = '123'",
				"viewRule": "missing = '123'",
				"createRule": "missing = '123'",
				"updateRule": "missing = '123'",
				"deleteRule": "missing = '123'"
			}`,
			[]string{"name", "schema", "listRule", "viewRule", "createRule", "updateRule", "deleteRule"},
		},
		// create failure - existing name
		{
			"",
			`{
				"name": "demo",
				"system": true,
				"schema": [
					{"name":"test","type":"text"}
				],
				"listRule": "test='123'",
				"viewRule": "test='123'",
				"createRule": "test='123'",
				"updateRule": "test='123'",
				"deleteRule": "test='123'"
			}`,
			[]string{"name"},
		},
		// create failure - existing internal table
		{
			"",
			`{
				"name": "_users",
				"schema": [
					{"name":"test","type":"text"}
				]
			}`,
			[]string{"name"},
		},
		// create failure - name starting with underscore
		{
			"",
			`{
				"name": "_test_new",
				"schema": [
					{"name":"test","type":"text"}
				]
			}`,
			[]string{"name"},
		},
		// create failure - duplicated field names (case insensitive)
		{
			"",
			`{
				"name": "test_new",
				"schema": [
					{"name":"test","type":"text"},
					{"name":"tESt","type":"text"}
				]
			}`,
			[]string{"schema"},
		},
		// create success
		{
			"",
			`{
				"name": "test_new",
				"system": true,
				"schema": [
					{"id":"a123456","name":"test1","type":"text"},
					{"id":"b123456","name":"test2","type":"email"}
				],
				"listRule": "test1='123'",
				"viewRule": "test1='123'",
				"createRule": "test1='123'",
				"updateRule": "test1='123'",
				"deleteRule": "test1='123'"
			}`,
			[]string{},
		},
		// update failure - changing field type
		{
			"test_new",
			`{
				"schema": [
					{"id":"a123456","name":"test1","type":"url"},
					{"id":"b123456","name":"test2","type":"bool"}
				]
			}`,
			[]string{"schema"},
		},
		// update failure - rename fields to existing field names (aka. reusing field names)
		{
			"test_new",
			`{
				"schema": [
					{"id":"a123456","name":"test2","type":"text"},
					{"id":"b123456","name":"test1","type":"email"}
				]
			}`,
			[]string{"schema"},
		},
		// update failure - existing name
		{
			"demo",
			`{"name": "demo2"}`,
			[]string{"name"},
		},
		// update failure - changing system collection
		{
			models.ProfileCollectionName,
			`{
				"name": "update",
				"system": false,
				"schema": [
					{"id":"koih1lqx","name":"userId","type":"text"}
				],
				"listRule": "userId = '123'",
				"viewRule": "userId = '123'",
				"createRule": "userId = '123'",
				"updateRule": "userId = '123'",
				"deleteRule": "userId = '123'"
			}`,
			[]string{"name", "system", "schema"},
		},
		// update failure - all fields
		{
			"demo",
			`{
				"name": "test ?!@#$",
				"system": true,
				"schema": [
					{"name":"","type":"text"}
				],
				"listRule": "missing = '123'",
				"viewRule": "missing = '123'",
				"createRule": "missing = '123'",
				"updateRule": "missing = '123'",
				"deleteRule": "missing = '123'"
			}`,
			[]string{"name", "system", "schema", "listRule", "viewRule", "createRule", "updateRule", "deleteRule"},
		},
		// update success - update all fields
		{
			"demo",
			`{
				"name": "demo_update",
				"schema": [
					{"id":"_2hlxbmp","name":"test","type":"text"}
				],
				"listRule": "test='123'",
				"viewRule": "test='123'",
				"createRule": "test='123'",
				"updateRule": "test='123'",
				"deleteRule": "test='123'"
			}`,
			[]string{},
		},
		// update failure - rename the schema field of the last updated collection
		// (fail due to filters old field references)
		{
			"demo_update",
			`{
				"schema": [
					{"id":"_2hlxbmp","name":"test_renamed","type":"text"}
				]
			}`,
			[]string{"listRule", "viewRule", "createRule", "updateRule", "deleteRule"},
		},
		// update success - rename the schema field of the last updated collection
		// (cleared filter references)
		{
			"demo_update",
			`{
				"schema": [
					{"id":"_2hlxbmp","name":"test_renamed","type":"text"}
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null
			}`,
			[]string{},
		},
		// update success - system collection
		{
			models.ProfileCollectionName,
			`{
				"listRule": "userId='123'",
				"viewRule": "userId='123'",
				"createRule": "userId='123'",
				"updateRule": "userId='123'",
				"deleteRule": "userId='123'"
			}`,
			[]string{},
		},
	}

	for i, s := range scenarios {
		collection := &models.Collection{}
		if s.existingName != "" {
			var err error
			collection, err = app.Dao().FindCollectionByNameOrId(s.existingName)
			if err != nil {
				t.Fatal(err)
			}
		}

		form := forms.NewCollectionUpsert(app, collection)

		// load data
		loadErr := json.Unmarshal([]byte(s.jsonData), form)
		if loadErr != nil {
			t.Errorf("(%d) Failed to load form data: %v", i, loadErr)
			continue
		}

		// parse errors
		result := form.Submit()
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("(%d) Failed to parse errors %v", i, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("(%d) Expected error keys %v, got %v", i, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("(%d) Missing expected error key %q in %v", i, k, errs)
			}
		}

		if len(s.expectedErrors) > 0 {
			continue
		}

		collection, _ = app.Dao().FindCollectionByNameOrId(form.Name)
		if collection == nil {
			t.Errorf("(%d) Expected to find collection %q, got nil", i, form.Name)
			continue
		}

		if form.Name != collection.Name {
			t.Errorf("(%d) Expected Name %q, got %q", i, collection.Name, form.Name)
		}

		if form.System != collection.System {
			t.Errorf("(%d) Expected System %v, got %v", i, collection.System, form.System)
		}

		if cast.ToString(form.ListRule) != cast.ToString(collection.ListRule) {
			t.Errorf("(%d) Expected ListRule %v, got %v", i, collection.ListRule, form.ListRule)
		}

		if cast.ToString(form.ViewRule) != cast.ToString(collection.ViewRule) {
			t.Errorf("(%d) Expected ViewRule %v, got %v", i, collection.ViewRule, form.ViewRule)
		}

		if cast.ToString(form.CreateRule) != cast.ToString(collection.CreateRule) {
			t.Errorf("(%d) Expected CreateRule %v, got %v", i, collection.CreateRule, form.CreateRule)
		}

		if cast.ToString(form.UpdateRule) != cast.ToString(collection.UpdateRule) {
			t.Errorf("(%d) Expected UpdateRule %v, got %v", i, collection.UpdateRule, form.UpdateRule)
		}

		if cast.ToString(form.DeleteRule) != cast.ToString(collection.DeleteRule) {
			t.Errorf("(%d) Expected DeleteRule %v, got %v", i, collection.DeleteRule, form.DeleteRule)
		}

		formSchema, _ := form.Schema.MarshalJSON()
		collectionSchema, _ := collection.Schema.MarshalJSON()
		if string(formSchema) != string(collectionSchema) {
			t.Errorf("(%d) Expected Schema %v, got %v", i, string(collectionSchema), string(formSchema))
		}
	}
}
