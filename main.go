package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/anacrolix/torrent"
)

// TorrentItem represents a torrent in our UI
type TorrentItem struct {
	Name       string
	Size       int64
	Downloaded int64
	Status     string
	Progress   float64
	Handle     *torrent.Torrent
}

func main() {
	// Create a new Fyne application
	a := app.New()
	w := a.NewWindow("Reed Torrent Client")
	w.Resize(fyne.NewSize(800, 600))

	// Create a torrent client
	cfg := torrent.NewDefaultClientConfig()
	// Set the download directory to the user's Downloads folder
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}
	cfg.DataDir = filepath.Join(homeDir, "Downloads", "ReedTorrent")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Error creating download directory: %v", err)
	}

	client, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating torrent client: %v", err)
	}
	defer client.Close()

	// Create a list of torrents
	torrentList := make(map[string]*TorrentItem)

	// Track the selected index
	selectedIndex := -1

	// Create the UI components
	magnetInput := widget.NewEntry()
	magnetInput.SetPlaceHolder("Enter magnet link or torrent URL")

	// Torrent list widget
	list := widget.NewList(
		func() int {
			return len(torrentList)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("Torrent Name"),
				container.NewHBox(
					widget.NewProgressBar(),
					widget.NewLabel("Status"),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// Get the torrent item
			var torrentItem *TorrentItem
			i := 0
			for _, t := range torrentList {
				if i == id {
					torrentItem = t
					break
				}
				i++
			}

			if torrentItem == nil {
				return
			}

			vbox := item.(*fyne.Container)
			nameLabel := vbox.Objects[0].(*widget.Label)
			hbox := vbox.Objects[1].(*fyne.Container)
			progressBar := hbox.Objects[0].(*widget.ProgressBar)
			statusLabel := hbox.Objects[1].(*widget.Label)

			nameLabel.SetText(torrentItem.Name)
			progressBar.Value = torrentItem.Progress
			statusLabel.SetText(torrentItem.Status)
		},
	)

	// Set up list selection
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
	}

	// Add torrent button
	addButton := widget.NewButtonWithIcon("Add Torrent", theme.ContentAddIcon(), func() {
		magnetLink := magnetInput.Text
		if magnetLink == "" {
			dialog.ShowError(fmt.Errorf("please enter a magnet link"), w)
			return
		}

		// Add the torrent
		t, err := client.AddMagnet(magnetLink)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error adding torrent: %v", err), w)
			return
		}

		// Wait for info
		go func() {
			<-t.GotInfo()

			// Add to our list
			torrentItem := &TorrentItem{
				Name:     t.Name(),
				Size:     t.Length(),
				Status:   "Downloading",
				Handle:   t,
				Progress: 0,
			}

			torrentList[t.InfoHash().String()] = torrentItem

			// Start downloading
			t.DownloadAll()

			// Update the UI
			list.Refresh()
		}()

		// Clear the input
		magnetInput.SetText("")
	})

	// Open torrent file button
	openButton := widget.NewButtonWithIcon("Open File", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}

			// Read the torrent file
			defer reader.Close()

			// Get the file path from the URI
			filePath := reader.URI().Path()

			// Add the torrent
			t, err := client.AddTorrentFromFile(filePath)
			if err != nil {
				dialog.ShowError(fmt.Errorf("error adding torrent: %v", err), w)
				return
			}

			// Wait for info
			go func() {
				<-t.GotInfo()

				// Add to our list
				torrentItem := &TorrentItem{
					Name:     t.Name(),
					Size:     t.Length(),
					Status:   "Downloading",
					Handle:   t,
					Progress: 0,
				}

				torrentList[t.InfoHash().String()] = torrentItem

				// Start downloading
				t.DownloadAll()

				// Update the UI
				list.Refresh()
			}()
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".torrent"}))
		fd.Show()
	})

	// Remove torrent button
	removeButton := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
		if selectedIndex < 0 {
			dialog.ShowInformation("Info", "Please select a torrent to remove", w)
			return
		}

		// Get the selected torrent
		var selectedTorrent *TorrentItem
		i := 0
		for _, t := range torrentList {
			if i == selectedIndex {
				selectedTorrent = t
				break
			}
			i++
		}

		if selectedTorrent == nil {
			return
		}

		// Remove the torrent
		selectedTorrent.Handle.Drop()
		delete(torrentList, selectedTorrent.Handle.InfoHash().String())

		// Update the UI
		list.Refresh()
	})

	// Create the main layout
	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				magnetInput,
				addButton,
				openButton,
			),
			container.NewHBox(
				removeButton,
			),
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		container.NewVBox(
			list,
		),
	)

	// Set the window content
	w.SetContent(content)

	// Start a goroutine to update the UI
	go func() {
		for {
			// Update torrent data (non-UI updates)
			for _, item := range torrentList {
				if item.Handle.Info() != nil {
					item.Downloaded = item.Handle.BytesCompleted()
					if item.Size > 0 {
						item.Progress = float64(item.Downloaded) / float64(item.Size)
					}

					// Update status
					if item.Progress >= 1.0 {
						item.Status = "Completed"
					} else if item.Handle.Seeding() {
						item.Status = "Seeding"
					} else {
						item.Status = fmt.Sprintf("Downloading (%.1f%%)", item.Progress*100)
					}
				}
			}

			// Use Async to safely update UI from a goroutine
			a.SendNotification(&fyne.Notification{
				Title:   "Reed",
				Content: "Refreshing UI",
			})

			// Trigger a UI refresh on the main thread
			list.Refresh()

			// Sleep for a bit
			time.Sleep(1 * time.Second)
		}
	}()

	// Show the window and run the app
	w.ShowAndRun()
}
