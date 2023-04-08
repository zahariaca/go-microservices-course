package toolbox

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Tools is the type for this package. Create a variable of this type, and you have access
// to all the exported methods with the receiver type *Tools.
type Tools struct {
	MaxJSONSize        int      // maximum size of JSON file we'll process
	MaxXMLSize         int      // maximum size of XML file we'll process
	MaxFileSize        int      // maximum size of uploaded files in bytes
	AllowedFileTypes   []string // allowed file types for upload (e.g. image/jpeg)
	AllowUnknownFields bool     // if set to true, allow unknown fields in JSON
}

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (t *Tools) ReadJson(w http.ResponseWriter, r *http.Request, data any) error {

	maxBytes := 1048576 // one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

func (t *Tools) WriteJson(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tools) ErrorJson(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	payload := JsonResponse{
		Error:   true,
		Message: err.Error(),
	}

	return t.WriteJson(w, statusCode, payload)
}
