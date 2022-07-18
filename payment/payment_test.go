package payment

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	// testValidDirectories contains a list of valid directory names
	testValidDirectories = []string{"20220717", "20120101"}
	// testInvalidDirectories contains a list of invalid directory names
	testInvalidDirectories = []string{"xyz", "2022_07_17", "2022", "202207", "202207177"}
	// testValidFiles contains a list of valid file names
	testValidFiles = []string{"010100.payments", "123000.payments"}
	// testInvalidFiles contains a list of invalid file names
	testInvalidFiles = []string{"01", "010100", ".payments", "a.payments", "aabbcc"}

	// testRawCSV contains sample payment input
	testRawCSV = `date,time,sequence,amount,comment
20220717,090000,211,500,payment2`
)

// TestNewWithBaseDir covers PaymentsService initialization
func TestNewWithBaseDir(t *testing.T) {
	t.Run("initialize with nonexistent directory", func(t *testing.T) {
		paymentsService, err := NewWithBaseDir("/tmp/nonexistent")
		if err == nil {
			t.Fatal("should error")
		}
		if paymentsService != nil {
			t.Fatal("paymentsService shouldn't be nil")
		}
	})
	t.Run("initialize with an existing directory", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("/tmp", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		paymentsService, err := NewWithBaseDir(tempDir)
		if err != nil {
			t.Fatal(err)
		}
		if paymentsService == nil {
			t.Fatal("paymentsServices shou;dn't be nil")
		}
	})
}

// TestParsePayments covers parsePayments functionality
func TestParsePayments(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	payments, err := paymentsService.parsePayments(strings.NewReader(testRawCSV))
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 1 {
		t.Fatalf("invalid payments length, got %d, expected %d", len(payments), 1)
	}
	if err := testValidatePayment(&payments[0]); err != nil {
		t.Fatal(err)
	}
}

// serviceWithTempDir is a helper that initializes PaymentsService with a temp dir
// The caller takes care of temp dir removal
func serviceWithTempDir() (*PaymentsService, string, error) {
	tempDir, err := ioutil.TempDir("/tmp", "test")
	if err != nil {
		return nil, "", err
	}
	paymentsService, err := NewWithBaseDir(tempDir)
	if err != nil {
		return nil, "", err
	}
	return paymentsService, tempDir, nil
}

// TestListDirectories is a basic test for ListDirectories functionality
func TestListDirectories(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	payments, err := paymentsService.ListDirectories()
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 0 {
		t.Fatal("invalid initial directory length, expected 0")
	}
	t.Run("list single valid directory", func(t *testing.T) {
		paymentsService, tempDir, err := serviceWithTempDir()
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		testDirectoryPath := filepath.Join(tempDir, "20220717")
		if os.Mkdir(testDirectoryPath, 0700); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(testDirectoryPath)
		payments, _ := paymentsService.ListDirectories()
		if len(payments) != 1 {
			t.Fatal("invalid directory length, expected 1")
		}
	})

	t.Run("list invalid directories", func(t *testing.T) {
		paymentsService, tempDir, err := serviceWithTempDir()
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		for _, invalidDir := range testInvalidDirectories {
			testDirectoryPath := filepath.Join(tempDir, invalidDir)
			if os.Mkdir(testDirectoryPath, 0700); err != nil {
				t.Fatal(err)
			}
			// The returned error is ignored here because ListDirectories only warns when invalid directories are used:
			payments, _ := paymentsService.ListDirectories()
			if len(payments) != 0 {

				t.Fatal("invalid directory length, should be 0")
			}
			os.Remove(testDirectoryPath)
		}
	})
}

// TestListPayments covers ListPayments functionality
func TestListPayments(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	testDirPath := filepath.Join(tempDir, "20220717")
	if err := os.Mkdir(testDirPath, 0700); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDirPath)
	testPaymentsFileName := "090030.payments"
	testFilePath := filepath.Join(testDirPath, testPaymentsFileName)
	testFile, err := os.Create(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer testFile.Close()
	payments, err := paymentsService.ListPayments("20220717")
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 1 {
		t.Fatalf("invalid payments length, got %d, expected %d", len(payments), 1)
	}
	if strings.Compare(payments[0], testPaymentsFileName) != 0 {
		t.Fatalf("invalid payments file name, got '%s', expected '%s'", payments[1], testPaymentsFileName)
	}
}

// TestGetPayments covers GetPayments functionality
func TestGetPayments(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	testDirPath := filepath.Join(tempDir, "20220717")
	if err := os.Mkdir(testDirPath, 0700); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDirPath)
	testPaymentsFileName := "090030.payments"
	testFilePath := filepath.Join(testDirPath, testPaymentsFileName)
	err = ioutil.WriteFile(testFilePath, []byte(testRawCSV), 0700)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFilePath)
	payments, err := paymentsService.GetPayments("20220717/090030.payments")
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 1 {
		t.Fatalf("invalid payments length, got %d, expected %d", len(payments), 1)
	}
	if err := testValidatePayment(&payments[0]); err != nil {
		t.Fatal(err)
	}
}

// TestValidateDirName covers validateDirName functionality using sample inputs
func TestValidateDirName(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	for _, d := range testInvalidDirectories {
		if err := paymentsService.validateDirName(d); err == nil {
			t.Fatalf("should error with dir name '%s'", d)
		}
	}
	for _, d := range testValidDirectories {
		if err := paymentsService.validateDirName(d); err != nil {
			t.Fatalf("shouldn accept dir name '%s'", d)
		}
	}
}

// TestValidateFileName covers validateFileName functionality using sample inputs
func TestValidateFileName(t *testing.T) {
	paymentsService, tempDir, err := serviceWithTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	for _, f := range testInvalidFiles {
		if err := paymentsService.validateFileName(f); err == nil {
			t.Fatalf("should error with filename '%s'", f)
		}
	}
	for _, f := range testValidFiles {
		if err := paymentsService.validateFileName(f); err != nil {
			t.Fatalf("should accept filename '%s'", f)
		}
	}
}

// testValidatePayment is a helper to validate the payment record that's contained in testRawCSV
func testValidatePayment(payment *Payment) error {
	expectedAsOfVal := 20220717090000
	if payment.AsOf != expectedAsOfVal {
		return fmt.Errorf("invalid asOf value, got %d, expected %d", payment.AsOf, expectedAsOfVal)
	}
	expectedSequence := 211
	if payment.Sequence != expectedSequence {
		return fmt.Errorf("invalid sequence value, got %d, expected %d", payment.Sequence, expectedSequence)
	}
	expectedAmount := 500
	if payment.Amount != expectedAmount {
		return fmt.Errorf("invalid amount value, got %d, expected %d", payment.Amount, expectedAmount)
	}
	expectedComment := "payment2"
	if strings.Compare(expectedComment, payment.Comment) != 0 {
		return fmt.Errorf("invalid comment value, got '%s', expected '%s'", payment.Comment, expectedComment)
	}
	return nil
}
