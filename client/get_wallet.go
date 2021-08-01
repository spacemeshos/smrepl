package client

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
)

func walkMatchX(root, pattern string, dirz bool) ([]string, error) {
	var matches []string
	if strings.HasSuffix(root, "/") {
		root = root[:len(root)-1]
	}
	files, err := ioutil.ReadDir(root)
	for _, info := range files {
		if err != nil {
			continue
		}
		if info.IsDir() {
			if dirz {
				matches = append(matches, root+"/"+info.Name())
			}
			continue
		}
		if dirz {
			continue
		}
		if matched, err := filepath.Match(pattern, filepath.Base(info.Name())); err != nil {
			continue
		} else if matched {
			matches = append(matches, root+"/"+info.Name())
		}
	}
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func walkMatch(root, pattern string) ([]string, error) {
	return walkMatchX(root, pattern, false)
}

func walkMatchDir(root string) ([]string, error) {
	return walkMatchX(root, "", true)
}

var thisDir string

func completer(d prompt.Document) []prompt.Suggest {
	parent := filepath.Dir(thisDir)
	s := []prompt.Suggest{
		{Text: parent, Description: "Parent Directory"},
	}
	jasonz, err := walkMatch(thisDir, "*.json")
	if err == nil {
		for _, fn := range jasonz {
			s = append(s, prompt.Suggest{Text: fn, Description: "JSON file"})
		}
	} else {
		log.Fatal("json walk", err)
	}
	filez, err := walkMatchDir(thisDir)
	if err == nil {
		for _, fn := range filez {
			if fn != thisDir {
				s = append(s, prompt.Suggest{Text: fn, Description: "sub directory"})
			}
		}
	} else {
		log.Fatal("dir walk", err)
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)

}

func (w *WalletBackend) getWallet() string {
	thisDir = w.workingDirectory
	for {
		t := prompt.Input(">", completer)
		if len(t) == 0 {
			return ""
		}

		fi, err := os.Lstat(t)
		if err != nil {
			fmt.Println("Error reading from wallet file. Press TAB and select another wallet file.")
			continue
		}

		if fi.IsDir() {
			thisDir = t
			continue
		}
		return t
	}
}
