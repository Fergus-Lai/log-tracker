package main

import (
	"time"
)

type ViewState uint

const (
	titleView ViewState = iota
	listView
	inputView
)

type model struct {
	state ViewState;
	lists listsModel;
	input inputModel;
	title titleModel;
	width  int;
    height int;
}

type listsModel struct {
	lists []listModel;
	selectedIndex uint;
}

type listModel struct {
	file File;
	filter Filter;
	cursor int
}

type inputModel struct {
	file File
}

type titleModel struct {
	choices []string
	selected int
	errorMessage string
}

type File struct {
	name string;
	folderPath string;
	fileNameString string;
	data []RowData;
	formatter string;
}

type RowData struct {
	level string;
	timeStamp time.Time;
	message string;
}

type Filter struct {
	searchString string;
	regexOn bool;
	level string;
	timeStampStart time.Time;
	timeStampEnd time.Time;
}