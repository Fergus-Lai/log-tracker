package main

import (
	"time"

	spinner "charm.land/bubbles/v2/spinner"
	textinput "charm.land/bubbles/v2/textinput"
)

type ViewState uint

const (
	titleView ViewState = iota
	listView
	inputView
	editView
)

type model struct {
	state   ViewState
	lists   listsModel
	input   inputModel
	title   titleModel
	edit    editModel
	width   int
	height  int
	files   []File
	spinner spinner.Model
}

type listsModel struct {
	selected      []bool
	selectedCount int
	focusedTab    int
	Filter        filterModel
	Cursor        int
	command       textinput.Model
}

type filterModel struct {
	regexOn bool
	inputs  []textinput.Model
}

type inputModel struct {
	focusIndex   int
	inputs       []textinput.Model
	choices      []string
	activeIndex  int
	errorMessage string
	inProgress   bool
}

type titleModel struct {
	choices      []string
	selected     int
	errorMessage string
}

type editModel struct {
	selectedIndex int
	editMode      bool
	input         inputModel
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

type dataLoadedMsg struct {
	files []File
	err   error
}
