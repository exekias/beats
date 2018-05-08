// +build !integration

package apiserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	mbtest "github.com/elastic/beats/metricbeat/mb/testing"

	"github.com/stretchr/testify/assert"
)

const testFile = "_meta/test/metrics"

func TestEventMapping(t *testing.T) {
	file, err := os.Open(testFile)
	assert.NoError(t, err, "cannot open test file "+testFile)

	body, err := ioutil.ReadAll(file)
	assert.NoError(t, err, "cannot read test file "+testFile)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")
		w.Write([]byte(body))
	}))

	server.Start()
	defer server.Close()

	config := map[string]interface{}{
		"module":     "kubernetes",
		"metricsets": []string{"apiserver"},
		"hosts":      []string{server.URL},
	}

	f := mbtest.NewEventsFetcher(t, config)

	events, err := f.Fetch()
	assert.NoError(t, err)

	fmt.Println(events)
	t.Fatal("foo")
}
