package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
)

const ObjectSeparator = "\n"
const FieldSeparator = "\t"

// The TextOutputWriter formats the CLI output as text meaning fields are separated
// by tabs and multiple elements by newline.
//
// It is used when the --output text parameter is provided.
// Example:
// foo1	bar1
// foo2	bar2
type TextOutputWriter struct {
	output      io.Writer
	transformer Transformer
}

func (w TextOutputWriter) sortKeys(value map[string]interface{}) []string {
	keys := []string{}
	for key, value := range value {
		if w.supportedValue(value) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func (w TextOutputWriter) supportedValue(value interface{}) bool {
	switch value.(type) {
	case float64, string, bool:
		return true
	}
	return false
}

func (w TextOutputWriter) writeValue(value interface{}) {
	if w.supportedValue(value) {
		switch v := value.(type) {
		case float64:
			_, _ = fmt.Fprint(w.output, strconv.FormatFloat(v, 'f', -1, 64))
		default:
			_, _ = fmt.Fprintf(w.output, "%v", value)
		}
	}
}

func (w TextOutputWriter) write(value interface{}, sortedBy []string) {
	switch result := value.(type) {
	case map[string]interface{}:
		if sortedBy == nil {
			sortedBy = w.sortKeys(result)
		}
		w.writeObject(result, sortedBy)
	case []interface{}:
		w.writeArray(result)
	default:
		w.writeValue(result)
		_, _ = fmt.Fprint(w.output, ObjectSeparator)
	}
}

func (w TextOutputWriter) collectObjectKeys(array []interface{}) []string {
	uniqueKeys := map[string]interface{}{}
	for _, row := range array {
		result, mapOk := row.(map[string]interface{})
		if mapOk {
			for key, value := range result {
				uniqueKeys[key] = value
			}
		}
	}
	return w.sortKeys(uniqueKeys)
}

func (w TextOutputWriter) writeRow(array []interface{}) {
	printTab := false
	for _, value := range array {
		if printTab {
			_, _ = fmt.Fprint(w.output, FieldSeparator)
		}
		printTab = true
		w.writeValue(value)
	}
	_, _ = fmt.Fprint(w.output, ObjectSeparator)
}

func (w TextOutputWriter) writeArray(array []interface{}) {
	keys := w.collectObjectKeys(array)
	for _, row := range array {
		switch result := row.(type) {
		case map[string]interface{}:
			w.write(result, keys)
		case []interface{}:
			w.writeRow(result)
		default:
			w.write(result, keys)
		}
	}
}

func (w TextOutputWriter) writeObject(obj map[string]interface{}, sortedBy []string) {
	values := []interface{}{}
	for _, key := range sortedBy {
		value, ok := obj[key]
		if !ok {
			value = ""
		}
		values = append(values, value)
	}
	w.writeRow(values)
}

func (w TextOutputWriter) writeBody(body []byte) error {
	var data interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		_, _ = fmt.Fprint(w.output, string(body))
		return nil
	}

	transformedResult, err := w.transformer.Execute(data)
	if err != nil {
		return err
	}
	w.write(transformedResult, nil)
	return nil
}

func (w TextOutputWriter) WriteResponse(response ResponseInfo) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 && response.StatusCode >= 400 {
		_, _ = fmt.Fprintf(w.output, "%s %s\n", response.Protocol, response.Status)
		return nil
	}
	return w.writeBody(body)
}

func NewTextOutputWriter(output io.Writer, transformer Transformer) *TextOutputWriter {
	return &TextOutputWriter{output, transformer}
}
