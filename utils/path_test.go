package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathExists(t *testing.T) {
	type args struct {
		p string
	}

	ex, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exPath := filepath.Dir(ex)
	t.Log(exPath)

	dir, _ := ioutil.TempDir(exPath, "pathtest")
	defer os.RemoveAll(dir)

	wantString := dir
	input := strings.Split(dir, string(filepath.Separator))
	inputString := input[len(input)-1]

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Path exists", args{p: inputString}, wantString, false},
		{"Path does not exist", args{p: "none"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathExists(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PathExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
