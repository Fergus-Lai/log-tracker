package main

import (
	"time"

	textinput "charm.land/bubbles/v2/textinput"
)

type ViewState uint

const (
	titleView ViewState = iota
	listView
	inputView
)

type model struct {
	state  ViewState
	lists  listsModel
	input  inputModel
	title  titleModel
	width  int
	height int
}

type listsModel struct {
	lists         []listModel
	selectedIndex uint
}

type listModel struct {
	File   File
	Filter Filter
	Cursor int
}

type inputModel struct {
	focusIndex   int
	inputs       []textinput.Model
	isSave       bool
	errorMessage string
	saving       bool
}

type titleModel struct {
	choices      []string
	selected     int
	errorMessage string
}

type File struct {
	Name           string `json:"name"`
	FolderPath     string `json:"folderPath"`
	FileNameString string `json:"fileNameMatch"`
	Formatter      string `json:"formatter"`
	Data           []RowData
}

type RowData struct {
	level     string
	timeStamp time.Time
	message   string
}

type Filter struct {
	searchString string
	regexOn      bool
	level        string
}

type dataLoadedMsg struct {
	lists []listModel
	err   error
}
