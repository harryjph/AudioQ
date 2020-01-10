package gui

import (
	"fmt"
	"github.com/harry1453/audioQ/constants"
	"github.com/harry1453/audioQ/project"
	"github.com/harry1453/go-common-file-dialog/cfd"
	"github.com/harry1453/go-common-file-dialog/cfdutil"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"os"
	"strconv"
)

func Initialize() {
	nameUpdateChannel := make(chan string)
	project.AddNameListener(nameUpdateChannel)
	settingsUpdateChannel := make(chan project.Settings)
	project.AddSettingsListener(settingsUpdateChannel)
	settingsStringUpdateChannel := make(chan string)
	go func() {
		for {
			settingsStringUpdateChannel <- strconv.Itoa(int((<-settingsUpdateChannel).BufferSize))
		}
	}()

	var cueName *walk.TextEdit
	var cueTable *walk.TableView
	cueTableModel := NewCueModel()

	project.AddCuesUpdateListener(func() {
		cueTableModel.ResetRows()
	})

	var window *walk.MainWindow

	exit, err := MainWindow{
		AssignTo: &window,
		Title:    "AudioQ " + constants.VERSION,
		Name:     "AudioQ " + constants.VERSION,
		Layout:   HBox{},
		Size: Size{
			Width:  100,
			Height: 100,
		},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Composite{
						Name:   "Control View",
						Layout: VBox{},
						Children: []Widget{
							TableView{
								AssignTo:         &cueTable,
								AlternatingRowBG: true,
								CheckBoxes:       true,
								ColumnsOrderable: true,
								MultiSelection:   true,
								Columns: []TableViewColumn{
									{Title: "#"},
									{Title: "Sel"},
									{Title: "Name"},
								},
								Model: cueTableModel,
							},
							Composite{
								Layout: HBox{Spacing: 5},
								Children: []Widget{
									PushButton{
										Text: "Play",
										OnClicked: func() {
											project.PlayNext() // TODO error
										},
									},
									PushButton{
										Text: "Stop",
										OnClicked: func() {
											project.StopPlaying()
										},
									},
									PushButton{
										Text: "Move Cue",
										OnClicked: func() {
											fromString, err := prompt(window, "Index From?")
											if err != nil || fromString == "" {
												fmt.Println("Error", err)
												return
											}
											from, err := strconv.Atoi(fromString)
											if err != nil {
												fmt.Println("Error", err)
												return
											}
											toString, err := prompt(window, "Index To?")
											if err != nil || toString == "" {
												fmt.Println("Error", err)
												return
											}
											to, err := strconv.Atoi(fromString)
											if err != nil {
												fmt.Println("Error", err)
												return
											}
											if err := project.MoveCue(from, to); err != nil {
												fmt.Println("Error", err)
												return
											}
										},
									},
								},
							},
						},
					},
					Composite{
						Name:   "Project View",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{Spacing: 1},
								Children: []Widget{
									PushButton{
										Text: "Open",
										OnClicked: func() {
											fileName, err := cfdutil.ShowOpenFileDialog(cfd.DialogConfig{
												Title: "Open AudioQ File",
												FileFilters: []cfd.FileFilter{
													{
														DisplayName: "AudioQ File (*.audioq)",
														Pattern:     "*.audioq",
													},
												},
											})
											if err != nil {
												log.Println("Error showing open file dialog:", err)
												return
											}
											if err := project.LoadProjectFile(fileName); err != nil {
												log.Println("Error loading file:", err)
											}
										},
									},
									PushButton{
										Text: "Save",
										OnClicked: func() {
											fileName, err := cfdutil.ShowSaveFileDialog(cfd.DialogConfig{
												Title: "Save AudioQ File",
												FileFilters: []cfd.FileFilter{
													{
														DisplayName: "AudioQ File (*.audioq)",
														Pattern:     "*.audioq",
													},
												},
											})
											if err != nil {
												log.Println("Error showing open file dialog:", err)
												return
											}
											if err := project.SaveProjectFile(fileName); err != nil {
												log.Println("Error saving file:", err)
											}
										},
									},
								},
							},
							setting("Project name", func(newName string) {
								project.SetName(newName)
							}, nameUpdateChannel),
							setting("Buffer Size", func(newBufferSize string) {
								n, err := strconv.Atoi(newBufferSize)
								if err != nil {
									log.Println("Failed to parse buffer size:", newBufferSize, err)
									return
								}
								project.SetSettings(project.Settings{BufferSize: uint(n)})
							}, settingsStringUpdateChannel),
							Composite{
								Layout: HBox{Spacing: 1},
								Children: []Widget{
									TextLabel{Text: "Cue Name:"},
									TextEdit{
										AssignTo: &cueName,
									},
									PushButton{
										Text: "Add Cue",
										OnClicked: func() {
											fileName, err := cfdutil.ShowOpenFileDialog(cfd.DialogConfig{
												Title: "Open Cue",
												FileFilters: []cfd.FileFilter{
													{
														DisplayName: "Audio Files (*.wav, *.flac, *.mp3, *.ogg",
														Pattern:     "*.wav;*.flac;*.mp3;*.ogg",
													},
												},
											})
											if err != nil {
												log.Println("Error showing open file dialog:", err)
												return
											}
											file, err := os.Open(fileName)
											if err != nil {
												log.Println("Error opening file:", err)
												return
											}
											if err := project.AddCue(cueName.Text(), fileName, file); err != nil {
												log.Println("Error adding cue:", err)
												return
											}
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}.Run()
	if err != nil {
		log.Println("GUI Error:", exit, err)
	}
}
