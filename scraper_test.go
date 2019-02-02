package main

import (
	"bytes"
	"testing"
)

func TestGetCleanedTableBodyDataEmpty(t *testing.T) {
	expected := []byte{}
	actual := getCleanedTableBodyData([]byte{})

	if !bytes.Equal(expected, actual) {
		t.Errorf("The expected output did not match the actual one. Expected: %v, Actual: %v", expected, actual)
	}
}

func TestGetCleanedTableBodyDataNoTableData(t *testing.T) {
	expected := []byte{}
	actual := getCleanedTableBodyData([]byte{'<', 'h', '1', '>', '<', '/', 'h', '1', '>'})

	if !bytes.Equal(expected, actual) {
		t.Errorf("The expected output did not match the actual one. Expected: %v, Actual: %v", expected, actual)
	}
}

func TestGetCleanedTableBodyDataNoTableBodyBuTableRow(t *testing.T) {
	expected := []byte{}
	actual := getCleanedTableBodyData([]byte{'<', 't', 'r', '>', '<', 't', 'd', '>', 't', 'e', 's', 't', '<', '/', 't', 'd', '>', '<', '/', 't', 'r', '>'})

	if !bytes.Equal(expected, actual) {
		t.Errorf("The expected output did not match the actual one. Expected: %v, Actual: %v", expected, actual)
	}
}

func TestGetCleanedTableBodyDataTableBodyClean(t *testing.T) {
	expected := []byte{'<', 't', 'b', 'o', 'd', 'y', '>', '<', 't', 'r', '>', '<', 't', 'd', '>', 't', 'e', 's', 't', '<', '/', 't', 'd', '>', '<', '/', 't', 'r', '>', '<', '/', 't', 'b', 'o', 'd', 'y', '>'}
	actual := getCleanedTableBodyData([]byte{'<', 't', 'b', 'o', 'd', 'y', '>', '<', 't', 'r', '>', '<', 't', 'd', '>', 't', 'e', 's', 't', '<', '/', 't', 'd', '>', '<', '/', 't', 'r', '>', '<', '/', 't', 'b', 'o', 'd', 'y', '>'})

	if !bytes.Equal(expected, actual) {
		t.Errorf("The expected output did not match the actual one. Expected: %v, Actual: %v", expected, actual)
	}
}

func TestGetCleanedTableBodyDataTableBodyDirty(t *testing.T) {
	expected := []byte{'<', 't', 'b', 'o', 'd', 'y', '>', '<', 't', 'r', '>', '<', 't', 'd', '>', 't', 'e', 's', 't', '<', '/', 't', 'd', '>', '<', '/', 't', 'r', '>', '<', '/', 't', 'b', 'o', 'd', 'y', '>'}
	actual := getCleanedTableBodyData([]byte{'<', 't', 'b', 'o', 'd', 'y', '>', '\n', '<', 't', 'r', '>', '<', 't', 'd', '>', 't', 'e', 's', 't', ',', '<', '/', 't', 'd', '>', '<', '/', 't', 'r', '>', '<', '/', 't', 'b', 'o', 'd', 'y', '>'})

	if !bytes.Equal(expected, actual) {
		t.Errorf("The expected output did not match the actual one. Expected: %v, Actual: %v", expected, actual)
	}
}
