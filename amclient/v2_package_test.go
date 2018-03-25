package amclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestPackage_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/v2beta/package/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		blob, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var expected = []byte(`{"name":"Foobar","type":"standard","accession":"","path":"PHV1aWQ+OjxwYXRoPg==","metadata_set_id":""}`)
		if !bytes.Equal(bytes.TrimSpace(blob), expected) {
			t.Fatal("path attribute does not have the expected value")
		}

		fmt.Fprint(w, `{"id": "096a284d-5067-4de0-a0a4-a684018cd6df"}`)
	})

	req := &PackageCreateRequest{
		Name: "Foobar",
		Type: "standard",
		Path: "<uuid>:<path>",
	}
	payload, _, _ := client.Package.Create(ctx, req)

	if want, got := "096a284d-5067-4de0-a0a4-a684018cd6df", payload.ID; want != got {
		t.Errorf("Package.Create() id: got %v, want %v", got, want)
	}
}
