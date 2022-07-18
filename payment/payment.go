package payment

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// These layouts are used to validate both file names and directory names:
	dateLayout = "20060102"
	timeLayout = "150405.payments"
)

// PaymentsService is the base building block of the payments service
type PaymentsService struct {
	BaseDir string
}

// Payment is the data structure used by the external representation format:
type Payment struct {
	// AsOf uses date+time format: 20220717063000
	AsOf     int    `json:"asOf"`
	Sequence int    `json:"sequence"`
	Amount   int    `json:"amount"`
	Comment  string `json:"comment,omitempty"`
}

// NewWithBaseDir initializes PaymentsService with a given base data directory (BaseDir):
func NewWithBaseDir(baseDir string) (*PaymentsService, error) {
	// Ensure it's possible to read the base directory:
	if _, err := os.ReadDir(baseDir); err != nil {
		return nil, err
	}
	return &PaymentsService{BaseDir: baseDir}, nil
}

// ListDirectories lists all directories available in the base data directory (BaseDir):
func (p *PaymentsService) ListDirectories() ([]string, error) {
	entries, err := os.ReadDir(p.BaseDir)
	if err != nil {
		return nil, err
	}
	directories := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		// Ensure the directory name is valid, print a warning and skip the entry if not:
		if err := p.validateDirName(name); err != nil {
			log.Println(err)
			continue
		}
		directories = append(directories, name)
	}
	return directories, nil
}

// ListPayments takes a given directory and lists its payment files:
func (p *PaymentsService) ListPayments(dir string) ([]string, error) {
	paymentsDirPath := filepath.Join(p.BaseDir, dir)
	entries, err := os.ReadDir(paymentsDirPath)
	if err != nil {
		return nil, err
	}
	payments := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		// Ensure the file name is valid, print a warning and skip the entry if not:
		if err := p.validateFileName(name); err != nil {
			log.Println(err)
			continue
		}
		payments = append(payments, name)
	}
	return payments, nil
}

// parsePayments is a helper that takes an io.Reader with CSV data
// and returns a list of payments ([]Payment)
func (p *PaymentsService) parsePayments(r io.Reader) ([]Payment, error) {
	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	payments := make([]Payment, 0)
	for i, row := range records {
		// Skip CSV header:
		if i == 0 {
			continue
		}
		// Parse and convert date and time to int:
		dateTimeStr := row[0] + row[1]
		dateTime, err := strconv.Atoi(dateTimeStr)
		if err != nil {
			log.Printf("Invalid date/time field in row %d: %s\n", i, err.Error())
			continue
		}
		// Parse sequence:
		sequence, err := strconv.Atoi(row[2])
		if err != nil {
			log.Printf("Invalid sequence field in row %d: %s\n", i, err.Error())
			continue
		}
		// Parse amount field:
		amount, err := strconv.Atoi(row[3])
		if err != nil {
			log.Printf("Invalid amount field in row %d: %s\n", i, err.Error())
			continue
		}
		// Parse comment (optional)
		comment := row[4]

		// Build payment object:
		payment := Payment{
			AsOf:     dateTime,
			Sequence: sequence,
			Amount:   amount,
			Comment:  comment,
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

// validateDirName validates an input string against the YYYYMMDD format:
func (p *PaymentsService) validateDirName(s string) error {
	_, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid directory name '%s': %s", s, err.Error())
	}
	return nil
}

// validateFileName validates an input string against the HHMMSS format:
func (p *PaymentsService) validateFileName(s string) error {
	_, err := time.Parse(timeLayout, s)
	if err != nil {
		return fmt.Errorf("invalid file name '%s': %s", s, err.Error())
	}
	return nil
}

// GetPayments parses a given file and returns its external representation format:
func (p *PaymentsService) GetPayments(path string) ([]Payment, error) {
	paymentsPath := filepath.Join(p.BaseDir, path)
	rawCSV, err := ioutil.ReadFile(paymentsPath)
	if err != nil {
		return nil, err
	}
	return p.parsePayments(bytes.NewReader(rawCSV))
}
