package file

import (
	"io/ioutil"
	"testing"
)

// TestHash tests our hashing function
func TestHash(t *testing.T) {

	type Test struct {
		input  string
		output string
	}

	tests := []Test{{input: "hello", output: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{input: "steve", output: "9ce5770b3bb4b2a1d59be2d97e34379cd192299f"},
	}

	for _, test := range tests {

		// Create a file with the given content
		tmpfile, err := ioutil.TempFile("", "example")
		if err != nil {
			t.Fatalf("create a temporary file failed")
		}

		// Write the input
		_, err = tmpfile.Write([]byte(test.input))
		if err != nil {
			t.Fatalf("error writing temporary file")
		}

		out := ""
		out, err = HashFile(tmpfile.Name())
		if err != nil {
			t.Fatalf("failed to hash file")
		}

		if out != test.output {
			t.Fatalf("invalid hash %s != %s", out, test.output)
		}

	}

	// Hashing a missing file should fail
	_, err := HashFile("/this/does/not/exist")
	if err == nil {
		t.Fatalf("should have seen an error, didn't")
	}
}
