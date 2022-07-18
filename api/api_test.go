package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matiasinsaurralde/product-services/payment"
)

var (
	testRawData = map[string]string{
		"20220717/090000.payments": `date,time,sequence,amount,comment
20220717,090000,211,500,payment2
20220717,090000,212,600,payment3`,
		"20220718/010101.payments": `date,time,sequence,amount,comment
20220718,010101,300,1500,payment4
20220718,010101,301,3000,payment5`,
	}

	testDesiredData = map[string][]payment.Payment{
		"20220717/090000.payments": {
			{AsOf: 20220717090000, Sequence: 211, Amount: 500, Comment: "payment2"},
			{AsOf: 20220717090000, Sequence: 212, Amount: 600, Comment: "payment3"},
		},
		"20220718/010101.payments": {
			{AsOf: 20220718010101, Sequence: 300, Amount: 1500, Comment: "payment4"},
			{AsOf: 20220718010101, Sequence: 301, Amount: 3000, Comment: "payment5"},
		},
	}

	testDirectories = []string{"20220717", "20220718"}
	testPaths       = map[string]string{
		"20220717": "090000.payments",
		"20220718": "010101.payments",
	}
)

func strInSlice(a string, s []string) bool {
	for _, b := range s {
		if b == a {
			return true
		}
	}
	return false
}
func TestHandler(t *testing.T) {
	// Initialize some test data:
	tempDir, err := ioutil.TempDir("/tmp", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	// Populate test directory:
	for fullPath, rawCSV := range testRawData {
		basePath := filepath.Dir(fullPath)                  // 20220717
		prefixedDirPath := filepath.Join(tempDir, basePath) // /tmpdir/20220717
		if err := os.Mkdir(prefixedDirPath, 0700); err != nil {
			t.Fatal(err)
		}
		prefixedFilePath := filepath.Join(tempDir, fullPath) // /tmpdir/20220717/090000.payments
		err := ioutil.WriteFile(prefixedFilePath, []byte(rawCSV), 0700)
		if err != nil {
			t.Fatal(err)
		}
	}

	h, err := NewHandler(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(h)

	t.Run("list directories", func(t *testing.T) {
		res, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		if res.Body == nil {
			t.Fatal("nil response body")
		}
		defer res.Body.Close()
		rawBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		var directories []string
		if err := json.Unmarshal(rawBody, &directories); err != nil {
			t.Fatal(err)
		}
		if len(directories) != 2 {
			t.Fatal("unexpected directory length")
		}
		// Check if API returned directories match what we expect:
		for _, d := range directories {
			if !strInSlice(d, testDirectories) {
				t.Fatal("unexpected directory")
			}
		}
	})
	t.Run("list payments", func(t *testing.T) {
		for d, paymentPath := range testPaths {
			url := fmt.Sprintf("%s/%s/", ts.URL, d) // http://server/000000/
			res, err := http.Get(url)
			if err != nil {
				t.Fatal(err)
			}
			if res.Body == nil {
				t.Fatal("nil response body")
			}
			defer res.Body.Close()
			rawBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			var paymentsList []string
			if err := json.Unmarshal(rawBody, &paymentsList); err != nil {
				t.Fatal(err)
			}

			// Check if API returned payments files match current testPaths value:
			if len(paymentsList) != 1 {
				t.Fatal("unexpected payments length")
			}
			if strings.Compare(paymentsList[0], paymentPath) != 0 {
				t.Fatal("invalid payment path")
			}
		}
	})
	t.Run("get payments", func(t *testing.T) {
		for d, paymentPath := range testPaths {
			url := fmt.Sprintf("%s/%s/%s", ts.URL, d, paymentPath) // http://server/000000/000000.payment
			res, err := http.Get(url)
			if err != nil {
				t.Fatal(err)
			}
			if res.Body == nil {
				t.Fatal("nil response body")
			}
			defer res.Body.Close()
			rawBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			var payments []payment.Payment
			if err := json.Unmarshal(rawBody, &payments); err != nil {
				t.Fatal(err)
			}
			// Grab unmarshaled desired data and compare values for every record
			filename := filepath.Join(d, paymentPath)
			desiredData := testDesiredData[filename]
			// Assert payments length:
			if len(payments) != len(desiredData) {
				t.Fatalf("invalid payments length, got %d, expected %d", len(payments), len(desiredData))
			}
			for _, p := range payments {
				var match bool
				var p2 payment.Payment
				// Iterate through desired data until a matching sequence is found
				for _, p2 = range desiredData {
					// We assume the sequence number is unique for each record on the payments file
					// so we use it to match API provided records with test data
					if p.Sequence != p2.Sequence {
						continue
					}
					match = true
					break
				}
				// match would signal if desired data doesn't contain a given API provided record:
				if !match {
					t.Fatal("no record matching this sequence")
				}
				// Compare fields:
				if p.Amount != p2.Amount {
					t.Fatalf("amount field doesn't match, got %d, expected %d", p.Amount, p2.Amount)
				}
				if p.AsOf != p2.AsOf {
					t.Fatalf("asOf field doesn't match, got %d, expected %d", p.AsOf, p2.AsOf)
				}
				if strings.Compare(p.Comment, p2.Comment) != 0 {
					t.Fatalf("comment field doesn't match, got '%s', expected '%s'", p.Comment, p2.Comment)
				}
			}
		}
	})

	defer ts.Close()
}
