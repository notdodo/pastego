package pegmatch_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/notdodo/pastego/pegmatch"
)

func TestPegmatchSimple(t *testing.T) {
	m := [2]string{"password", "quake && ~earthquake"}
	countMatches := 0
	correctMatches := 2
	pegmatch.PasteContentString = "my password is: quake"
	for _, mtch := range m {
		mtch = strings.TrimSpace(mtch)
		got, err := pegmatch.ParseReader("", bytes.NewBufferString(mtch))
		if err == nil && got.(bool) {
			countMatches++
		}
	}
	if countMatches != correctMatches {
		t.Error("failed")
	}
}

func TestPegmatchMedium(t *testing.T) {
	m := [1]string{"password && ~(include || java)"}
	countMatches := 0
	correctMatches := 0
	pegmatch.PasteContentString = "my password is: java"
	for _, mtch := range m {
		mtch = strings.TrimSpace(mtch)
		got, err := pegmatch.ParseReader("", bytes.NewBufferString(mtch))
		if err == nil && got.(bool) {
			countMatches++
		}
	}
	if countMatches != correctMatches {
		t.Error("failed")
	}
}

func TestPegmatchHard(t *testing.T) {
	m := [2]string{"quake && ~earthquake", "php && ~(sudo || Linux || '<body>')"}
	countMatches := 0
	correctMatches := 2
	pegmatch.PasteContentString = "quakelive was good"
	for _, mtch := range m {
		mtch = strings.TrimSpace(mtch)
		got, err := pegmatch.ParseReader("", bytes.NewBufferString(mtch))
		if err == nil && got.(bool) {
			countMatches++
		}
	}
	pegmatch.PasteContentString = `
		<?php 
			echo '<input type="button" onclick="alert(\'OMG!\')"/>';
		?>`
	for _, mtch := range m {
		mtch = strings.TrimSpace(mtch)
		got, err := pegmatch.ParseReader("", bytes.NewBufferString(mtch))
		if err == nil && got.(bool) {
			countMatches++
		}
	}

	if countMatches != correctMatches {
		t.Error("failed")
	}
}
