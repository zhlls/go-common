package utils

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

var JsonSchemaError = errors.New("json schema error")

func JsonSchema(schema string, doc []byte) error {

	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(doc)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	} else {
		for _, desc := range result.Errors() {
			return errors.New(desc.String())
		}
	}

	return JsonSchemaError
}
