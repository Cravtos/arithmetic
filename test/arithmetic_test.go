package test

import (
	"github.com/cravtos/arithmetic/internal/pkg/arithmetic"
	"github.com/cravtos/arithmetic/internal/pkg/helpers"
	"io/ioutil"
	"os"
	"testing"
)

// TestArithmetic encodes and decodes every file in test/testdata, and compares results to originals.
func TestArithmetic(t *testing.T) {
	testFiles, err := ioutil.ReadDir("./testdata")
	if err != nil {
		t.Errorf("got error while getting testdata: %v\n", err)
		t.FailNow()
	}

	for _, file := range testFiles {
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			t.Parallel()
			equal, err := cmpEncodeAndDecode(t, "./testdata/"+file.Name())
			if err != nil {
				t.Errorf("got error while encoding or decoding testdata: %v\n", err)
			}
			if equal != true {
				t.Error("original and decoded files are not equal")
			}
		})
	}
}

// cmpEncodeAndDecode encodes and decodes files, and compares original to decoded one.
func cmpEncodeAndDecode(t *testing.T, file string) (equal bool, err error) {
	// Create temp directory for resulting files
	tmp := t.TempDir()

	origName := file
	origFile, err := os.Open(origName)
	if err != nil {
		return false, err
	}
	defer origFile.Close()

	// Create file for encoded data
	encName := tmp + ".encoded"
	encFile, err := os.Create(encName)
	if err != nil {
		return false, err
	}
	defer encFile.Close()
	defer os.Remove(encName)

	err = arithmetic.Encode(origFile, encFile)
	if err != nil {
		return false, err
	}

	// Seek to beginning (since it will be used again to be decoded)
	_, err = encFile.Seek(0, 0)
	if err != nil {
		return false, err
	}

	// Create file for decoded data
	decName := tmp + ".decoded"
	decFile, err := os.Create(decName)
	if err != nil {
		return false, err
	}
	defer decFile.Close()
	defer os.Remove(decName)

	err = arithmetic.Decode(encFile, decFile)
	if err != nil {
		return false, err
	}

	// Seek files to beginning for comparing
	_, err = origFile.Seek(0, 0)
	if err != nil {
		return false, err
	}

	_, err = decFile.Seek(0, 0)
	if err != nil {
		return false, err
	}

	equal, err = helpers.CompareFiles(origFile, decFile)
	if err != nil {
		return false, err
	}
	return equal, nil
}
