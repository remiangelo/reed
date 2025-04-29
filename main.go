package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/anacrolix/torrent"
)

// TorrentItem represents a torrent in our UI
type TorrentItem struct {
	Name         string
	Size         int64
	Downloaded   int64
	Status       string
	Progress     float64
	Handle       *torrent.Torrent
	DownloadRate int64      // Download rate in bytes per second
	UploadRate   int64      // Upload rate in bytes per second
	Peers        int        // Number of connected peers
	Seeds        int        // Number of connected seeds
	AddedAt      time.Time  // When the torrent was added
	LastUpdate   time.Time  // Last time stats were updated
	Files        []FileInfo // Information about files in the torrent
	FileCount    int        // Number of files in the torrent
	ETA          string     // Estimated time to completion
}

// FileInfo represents a file within a torrent
type FileInfo struct {
	Path     string
	Size     int64
	Progress float64
}

// HumanReadableSize converts bytes to a human-readable string
func HumanReadableSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	var size float64
	var unit string

	switch {
	case bytes >= TB:
		size = float64(bytes) / TB
		unit = "TB"
	case bytes >= GB:
		size = float64(bytes) / GB
		unit = "GB"
	case bytes >= MB:
		size = float64(bytes) / MB
		unit = "MB"
	case bytes >= KB:
		size = float64(bytes) / KB
		unit = "KB"
	default:
		size = float64(bytes)
		unit = "B"
	}

	return fmt.Sprintf("%.2f %s", size, unit)
}

// HumanReadableRate converts bytes/second to a human-readable string
func HumanReadableRate(bytesPerSec int64) string {
	if bytesPerSec == 0 {
		return "0 B/s"
	}
	return HumanReadableSize(bytesPerSec) + "/s"
}

func main() {
	// Create a new Fyne application with ID
	a := app.NewWithID("com.github.reed.torrentclient")
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

	// Helper function to validate torrent items and clean up invalid ones
	validateTorrents := func() {
		// Find torrents that have nil handles or other issues
		invalidTorrents := make([]string, 0)
		for hash, item := range torrentList {
			if item == nil || item.Handle == nil {
				invalidTorrents = append(invalidTorrents, hash)
			}
		}

		// Remove invalid torrents
		for _, hash := range invalidTorrents {
			delete(torrentList, hash)
		}
	}

	// Create the UI components
	magnetInput := widget.NewEntry()
	magnetInput.SetPlaceHolder("Enter magnet link or torrent URL")

	// Variable to reference the add torrent dialog
	var addTorrentDialog dialog.Dialog

	// Torrent list widget
	list := widget.NewList(
		func() int {
			return len(torrentList)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				container.NewHBox(
					widget.NewIcon(theme.FileIcon()),
					widget.NewLabel("Torrent Name"),
				),
				widget.NewProgressBar(),
				container.NewHBox(
					widget.NewLabel("Status:"),
					widget.NewLabel("Status"),
					widget.NewSeparator(),
					widget.NewLabel("Size:"),
					widget.NewLabel("Size"),
					widget.NewSeparator(),
					widget.NewLabel("Speed:"),
					widget.NewLabel("Speed"),
				),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// Convert the map to a slice for indexed access
			torrents := make([]*TorrentItem, 0, len(torrentList))
			for _, t := range torrentList {
				torrents = append(torrents, t)
			}

			// Safety check for index bounds
			if int(id) >= len(torrents) {
				return
			}

			// Get the torrent item at this index
			torrentItem := torrents[id]
			if torrentItem == nil {
				return
			}

			// Safe type assertions with fallbacks
			vbox, ok := item.(*fyne.Container)
			if !ok || len(vbox.Objects) < 3 {
				return
			}

			// Top row with icon and name
			hbox, ok := vbox.Objects[0].(*fyne.Container)
			if !ok || len(hbox.Objects) < 2 {
				return
			}

			nameLabel, ok := hbox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			// Progress bar
			progressBar, ok := vbox.Objects[1].(*widget.ProgressBar)
			if !ok {
				return
			}

			// Bottom row with stats
			statsBox, ok := vbox.Objects[2].(*fyne.Container)
			if !ok || len(statsBox.Objects) < 8 {
				return
			}

			statusLabel, ok := statsBox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			sizeLabel, ok := statsBox.Objects[4].(*widget.Label)
			if !ok {
				return
			}

			speedLabel, ok := statsBox.Objects[7].(*widget.Label)
			if !ok {
				return
			}

			// Set values safely
			nameLabel.SetText(torrentItem.Name)
			progressBar.Value = torrentItem.Progress
			statusLabel.SetText(torrentItem.Status)
			sizeLabel.SetText(HumanReadableSize(torrentItem.Size))

			if torrentItem.DownloadRate > 0 {
				speedLabel.SetText(HumanReadableRate(torrentItem.DownloadRate))
			} else if torrentItem.UploadRate > 0 {
				speedLabel.SetText("↑ " + HumanReadableRate(torrentItem.UploadRate))
			} else {
				speedLabel.SetText("-")
			}
		},
	)

	// Set up list selection
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
	}

	// Status bar for the bottom of the window (declared here so it can be accessed in the goroutine)
	statusBar := container.NewHBox(
		widget.NewLabel("Status: Ready"),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Download Directory: %s", cfg.DataDir)),
	)

	// Create a detail panel for the selected torrent
	var detailsContainer *fyne.Container
	detailsContainer = container.NewVBox(
		widget.NewLabel("No torrent selected"),
	)

	// Function to update the details panel will be defined later in the code
	var updateDetailsPanel func()

	// Create a toolbar with action buttons
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			// Create a larger, more functional add torrent dialog

			// Create a tab container for different ways to add torrents
			magnetInput.SetPlaceHolder("Enter magnet link or torrent URL")
			magnetInput.SetText("")

			// Create a multi-line text area for batch adding magnet links
			batchInput := widget.NewMultiLineEntry()
			batchInput.SetPlaceHolder("Enter multiple magnet links, one per line")

			addButton := widget.NewButton("Add Torrent", func() {
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

					// Create a standardized torrent item
					now := time.Now()
					torrentItem := &TorrentItem{
						Name:         t.Name(),
						Size:         t.Length(),
						Status:       "Downloading",
						Handle:       t,
						Progress:     0,
						Downloaded:   0,
						AddedAt:      now,
						LastUpdate:   now,
						DownloadRate: 0,
						UploadRate:   0,
						Peers:        0,
						Seeds:        0,
						FileCount:    len(t.Info().Files),
						ETA:          "Calculating...",
						Files:        []FileInfo{},
					}

					// Add to our list
					torrentList[t.InfoHash().String()] = torrentItem

					// Start downloading
					t.DownloadAll()

					// Update the UI safely from goroutine
					fyne.Do(func() {
						list.Refresh()
						updateDetailsPanel()
					})
				}()

				// Clear the input and close dialog
				magnetInput.SetText("")
				addTorrentDialog.Hide()
			})

			addBatchButton := widget.NewButton("Add All", func() {
				// Get all lines from the batch input
				magnetLinks := batchInput.Text
				if magnetLinks == "" {
					dialog.ShowError(fmt.Errorf("please enter at least one magnet link"), w)
					return
				}

				// Split by newlines
				links := strings.Split(magnetLinks, "\n")
				addedCount := 0

				for _, link := range links {
					link = strings.TrimSpace(link)
					if link == "" {
						continue
					}

					// Add each torrent
					t, err := client.AddMagnet(link)
					if err != nil {
						log.Printf("Error adding torrent: %v", err)
						continue
					}

					// Process in background
					go func(torrent *torrent.Torrent) {
						<-torrent.GotInfo()

						// Create a standardized torrent item
						now := time.Now()
						torrentItem := &TorrentItem{
							Name:         torrent.Name(),
							Size:         torrent.Length(),
							Status:       "Downloading",
							Handle:       torrent,
							Progress:     0,
							Downloaded:   0,
							AddedAt:      now,
							LastUpdate:   now,
							DownloadRate: 0,
							UploadRate:   0,
							Peers:        0,
							Seeds:        0,
							FileCount:    len(torrent.Info().Files),
							ETA:          "Calculating...",
							Files:        []FileInfo{},
						}

						torrentList[torrent.InfoHash().String()] = torrentItem

						// Start downloading
						torrent.DownloadAll()

						// Update the UI safely from goroutine
						fyne.Do(func() {
							list.Refresh()
							updateDetailsPanel()
						})
					}(t)

					addedCount++
				}

				// Show success message
				if addedCount > 0 {
					dialog.ShowInformation("Torrents Added", fmt.Sprintf("Added %d torrent(s).", addedCount), w)
				}

				// Clear the input and close dialog
				batchInput.SetText("")
				addTorrentDialog.Hide()
			})

			// Create tabs for different ways to add torrents
			tabs := container.NewAppTabs(
				container.NewTabItem("Magnet Link", container.NewVBox(
					widget.NewLabel("Enter magnet link or torrent URL:"),
					magnetInput,
					container.NewHBox(
						layout.NewSpacer(),
						widget.NewButton("Clear", func() {
							magnetInput.SetText("")
						}),
						addButton,
					),
				)),
				container.NewTabItem("Batch Add", container.NewVBox(
					widget.NewLabel("Enter multiple magnet links (one per line):"),
					container.NewVScroll(batchInput),
					container.NewHBox(
						layout.NewSpacer(),
						widget.NewButton("Clear", func() {
							batchInput.SetText("")
						}),
						addBatchButton,
					),
				)),
			)

			// Create dialog content
			dialogContent := container.NewVBox(
				tabs,
			)

			// Set minimum size for the dialog
			dialogContent.Resize(fyne.NewSize(500, 300))

			// Create and show dialog
			addTorrentDialog = dialog.NewCustom("Add Torrent", "Cancel", dialogContent, w)
			addTorrentDialog.Resize(fyne.NewSize(500, 300))
			addTorrentDialog.Show()
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
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

					// Create a standardized torrent item
					now := time.Now()
					torrentItem := &TorrentItem{
						Name:         t.Name(),
						Size:         t.Length(),
						Status:       "Downloading",
						Handle:       t,
						Progress:     0,
						Downloaded:   0,
						AddedAt:      now,
						LastUpdate:   now,
						DownloadRate: 0,
						UploadRate:   0,
						Peers:        0,
						Seeds:        0,
						FileCount:    len(t.Info().Files),
						ETA:          "Calculating...",
						Files:        []FileInfo{},
					}

					torrentList[t.InfoHash().String()] = torrentItem

					// Start downloading
					t.DownloadAll()

					// Update the UI safely from goroutine
					fyne.Do(func() {
						list.Refresh()
						updateDetailsPanel()
					})
				}()
			}, w)
			fd.SetFilter(storage.NewExtensionFileFilter([]string{".torrent"}))
			fd.Show()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			if selectedIndex < 0 {
				dialog.ShowInformation("Info", "Please select a torrent to remove", w)
				return
			}

			// Get the selected torrent safely using a slice
			torrents := make([]*TorrentItem, 0, len(torrentList))
			for _, t := range torrentList {
				torrents = append(torrents, t)
			}

			// Check index bounds
			if selectedIndex >= len(torrents) {
				dialog.ShowError(fmt.Errorf("invalid torrent selection"), w)
				return
			}

			// Get the selected torrent
			selectedTorrent := torrents[selectedIndex]

			// Validate torrent
			if selectedTorrent == nil {
				dialog.ShowError(fmt.Errorf("selected torrent is invalid"), w)
				return
			}

			// Validate handle
			if selectedTorrent.Handle == nil {
				dialog.ShowError(fmt.Errorf("torrent handle is invalid"), w)
				// Clean up the invalid torrent
				for hash, t := range torrentList {
					if t == selectedTorrent {
						delete(torrentList, hash)
						break
					}
				}
				list.Refresh()
				selectedIndex = -1
				return
			}

			// Show confirmation dialog
			confirmDialog := dialog.NewConfirm(
				"Remove Torrent",
				fmt.Sprintf("Are you sure you want to remove '%s'?", selectedTorrent.Name),
				func(confirmed bool) {
					if confirmed {
						// Get hash before dropping the torrent (with safety check)
						var hash string
						if selectedTorrent.Handle != nil {
							hash = selectedTorrent.Handle.InfoHash().String()

							// Drop the torrent
							selectedTorrent.Handle.Drop()
						} else {
							// If handle is nil, search for this torrent in the map to remove it
							for h, t := range torrentList {
								if t == selectedTorrent {
									hash = h
									break
								}
							}
						}

						// Remove from our list
						delete(torrentList, hash)

						// Update the UI
						list.Refresh()
						selectedIndex = -1

						// Update the details panel to show "No torrent selected"
						updateDetailsPanel()

						// Validate torrent list
						validateTorrents()
					}
				}, w)
			confirmDialog.Show()
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			// Show settings dialog
			dialog.ShowInformation("Settings", "Settings dialog will be implemented in a future update.", w)
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			dialog.ShowInformation("About Reed Torrent Client",
				"Reed Torrent Client v1.0.0\n\nA lightweight torrent client built with Go using the anacrolix/torrent library and Fyne for the UI.", w)
		}),
	)

	// The status bar is already declared above so we don't need to redeclare it here

	// Function to update the details panel
	updateDetailsPanel = func() {
		// Clear the container
		detailsContainer.Objects = nil

		if selectedIndex < 0 {
			detailsContainer.Add(widget.NewLabel("No torrent selected"))
			detailsContainer.Refresh()
			return
		}

		// Get the selected torrent safely
		var selectedTorrent *TorrentItem

		if selectedIndex >= 0 {
			// Convert map to a slice for indexed access
			torrents := make([]*TorrentItem, 0, len(torrentList))
			for _, t := range torrentList {
				torrents = append(torrents, t)
			}

			// Only access the slice if the index is valid
			if selectedIndex < len(torrents) {
				selectedTorrent = torrents[selectedIndex]
			}
		}

		if selectedTorrent == nil {
			detailsContainer.Add(widget.NewLabel("Torrent not found or none selected"))
			detailsContainer.Refresh()
			return
		}

		// Additional safety check
		if selectedTorrent.Handle == nil || selectedTorrent.Handle.Info() == nil {
			detailsContainer.Add(widget.NewLabel("Torrent information not available yet"))
			detailsContainer.Refresh()
			return
		}

		// Add torrent information to the details panel
		detailsContainer.Add(widget.NewLabelWithStyle(
			selectedTorrent.Name,
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		))

		// Create a more detailed info form
		infoForm := widget.NewForm(
			widget.NewFormItem("Status", widget.NewLabel(selectedTorrent.Status)),
			widget.NewFormItem("Size", widget.NewLabel(HumanReadableSize(selectedTorrent.Size))),
			widget.NewFormItem("Downloaded", widget.NewLabel(HumanReadableSize(selectedTorrent.Downloaded))),
			widget.NewFormItem("Progress", widget.NewLabel(fmt.Sprintf("%.1f%%", selectedTorrent.Progress*100))),
			widget.NewFormItem("Download Speed", widget.NewLabel(HumanReadableRate(selectedTorrent.DownloadRate))),
			widget.NewFormItem("Upload Speed", widget.NewLabel(HumanReadableRate(selectedTorrent.UploadRate))),
			widget.NewFormItem("Peers", widget.NewLabel(fmt.Sprintf("%d", selectedTorrent.Peers))),
		)

		// Add ETA if downloading
		if selectedTorrent.Progress < 1.0 && selectedTorrent.DownloadRate > 0 {
			infoForm.Append("ETA", widget.NewLabel(selectedTorrent.ETA))
		}

		// Add metadata info
		infoForm.Append("Added", widget.NewLabel(selectedTorrent.AddedAt.Format("2006-01-02 15:04:05")))

		// Calculate and show data transferred since added
		if selectedTorrent.Downloaded > 0 {
			infoForm.Append("Data Transferred", widget.NewLabel(HumanReadableSize(selectedTorrent.Downloaded)))
		}
		detailsContainer.Add(infoForm)

		// Actions for this torrent
		actionsContainer := container.NewHBox(
			widget.NewButton("Pause/Resume", func() {
				// Toggle pause/resume logic will be added later
				dialog.ShowInformation("Not Implemented", "Pause/Resume functionality will be added soon.", w)
			}),
			widget.NewButton("Open Folder", func() {
				// Open the download folder for this torrent
				dialog.ShowInformation("Open Folder", "This will open the folder containing the downloaded files.", w)
				// The actual implementation would be platform-specific
			}),
		)
		detailsContainer.Add(actionsContainer)

		// If the torrent has files, show them
		// Use multiple safety checks to prevent nil pointer dereferences
		if selectedTorrent != nil && selectedTorrent.Handle != nil &&
			selectedTorrent.Handle.Info() != nil &&
			len(selectedTorrent.Handle.Info().Files) > 0 {

			detailsContainer.Add(widget.NewLabelWithStyle("Files:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

			filesList := widget.NewList(
				func() int {
					// Double-check that info is still available (could change between renders)
					if selectedTorrent != nil && selectedTorrent.Handle != nil &&
						selectedTorrent.Handle.Info() != nil {
						return len(selectedTorrent.Handle.Info().Files)
					}
					return 0
				},
				func() fyne.CanvasObject {
					return container.NewHBox(
						widget.NewIcon(theme.FileIcon()),
						widget.NewLabel("Filename"),
						widget.NewProgressBar(),
						widget.NewLabel("Size"),
					)
				},
				func(id widget.ListItemID, obj fyne.CanvasObject) {
					// Safety checks
					if selectedTorrent == nil || selectedTorrent.Handle == nil ||
						selectedTorrent.Handle.Info() == nil ||
						int(id) >= len(selectedTorrent.Handle.Info().Files) {
						return
					}

					file := selectedTorrent.Handle.Info().Files[id]

					hbox := obj.(*fyne.Container)
					filenameLabel := hbox.Objects[1].(*widget.Label)
					progressBar := hbox.Objects[2].(*widget.ProgressBar)
					sizeLabel := hbox.Objects[3].(*widget.Label)

					// Get the filename from the path - file.Path is a slice of strings
					if len(file.Path) > 0 {
						// Use the last component as the filename
						filenameLabel.SetText(file.Path[len(file.Path)-1])
					} else {
						filenameLabel.SetText("Unknown file")
					}
					sizeLabel.SetText(HumanReadableSize(file.Length))
					progressBar.Value = selectedTorrent.Progress
				},
			)

			// Wrap the files list in a scroll container with fixed height
			filesScroll := container.NewVScroll(filesList)
			filesScroll.SetMinSize(fyne.NewSize(0, 150))
			detailsContainer.Add(filesScroll)
		} else if selectedTorrent.Handle.Info() != nil {
			// Single file torrent
			detailsContainer.Add(widget.NewLabelWithStyle("Single File:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			detailsContainer.Add(widget.NewLabel(selectedTorrent.Name))
		}

		detailsContainer.Refresh()
	}

	// Set up list selection to update the details panel - this overrides the previous OnSelected
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
		updateDetailsPanel()
	}

	// Create a split container with the list on the left and details on the right
	splitContainer := container.NewHSplit(
		container.NewVBox(list),
		container.NewScroll(detailsContainer),
	)
	splitContainer.Offset = 0.7 // 70% of space for the list, 30% for details

	// Create the main layout with the toolbar at the top
	content := container.NewBorder(
		container.NewVBox(
			toolbar,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			statusBar,
		),
		nil,
		nil,
		splitContainer,
	)

	// Set the window content
	w.SetContent(content)

	// Start a goroutine to update the UI
	go func() {
		// Maps to track previous download/upload byte counts
		prevDownloaded := make(map[string]int64)
		prevUploaded := make(map[string]int64)

		for {
			// First validate all torrents to remove any invalid ones
			validateTorrents()

			// Map to track newly completed torrents for notifications
			newlyCompleted := make(map[string]bool)

			// Update torrent data (non-UI updates)
			for hash, item := range torrentList {
				// Skip invalid torrents
				if item == nil || item.Handle == nil {
					continue
				}

				// Skip torrents without info yet
				if item.Handle.Info() == nil {
					continue
				}

				// Get current timestamp
				now := time.Now()

				// Whether this was previously marked as completed
				wasCompleted := item.Status == "Completed"

				// Update downloaded bytes
				currentBytes := item.Handle.BytesCompleted()
				previousBytes := item.Downloaded // Store for notification check
				item.Downloaded = currentBytes

				// Calculate download rate safely
				if prev, ok := prevDownloaded[hash]; ok {
					// Calculate time difference since last update
					timeDiffSec := now.Sub(item.LastUpdate).Seconds()
					if timeDiffSec > 0 {
						// Calculate and update download rate (bytes/second)
						byteDiff := currentBytes - prev
						if byteDiff >= 0 { // Ensure non-negative
							item.DownloadRate = int64(float64(byteDiff) / timeDiffSec)
						}
					}
				}
				// Store current bytes for next rate calculation
				prevDownloaded[hash] = currentBytes

				// Calculate upload rate (simplified version)
				// Note: In a real app, we'd track actual bytes uploaded
				currentUploaded := item.Handle.BytesCompleted()
				if prev, ok := prevUploaded[hash]; ok && prev > 0 {
					// Use different variable to avoid shadowing
					uploadTimeDiff := now.Sub(item.LastUpdate).Seconds()
					if uploadTimeDiff > 0 {
						// Calculate rate safely
						byteDiff := currentUploaded - prev
						if byteDiff >= 0 { // Ensure non-negative
							item.UploadRate = int64(float64(byteDiff) / uploadTimeDiff)
						}
					}
				}
				// Store current upload bytes for next calculation
				prevUploaded[hash] = currentUploaded

				// Update progress percentage
				if item.Size > 0 {
					item.Progress = float64(item.Downloaded) / float64(item.Size)
					// Cap progress at 100%
					if item.Progress > 1.0 {
						item.Progress = 1.0
					}
				}

				// Update status based on download progress
				if item.Progress >= 1.0 {
					item.Status = "Completed"
					item.ETA = ""

					// Check if this torrent was just completed
					if !wasCompleted && previousBytes < item.Size && currentBytes >= item.Size {
						newlyCompleted[hash] = true
					}
				} else if item.Handle.Seeding() {
					item.Status = "Seeding"
					item.ETA = ""
				} else {
					item.Status = fmt.Sprintf("Downloading (%.1f%%)", item.Progress*100)

					// Calculate ETA if downloading at a reasonable rate
					if item.DownloadRate > 1024 { // Only if downloading faster than 1 KB/s
						remainingBytes := item.Size - item.Downloaded
						secondsRemaining := float64(remainingBytes) / float64(item.DownloadRate)

						// Format ETA based on time remaining
						if secondsRemaining < 60 {
							item.ETA = fmt.Sprintf("%.0f sec", secondsRemaining)
						} else if secondsRemaining < 3600 {
							item.ETA = fmt.Sprintf("%.1f min", secondsRemaining/60)
						} else if secondsRemaining < 86400 {
							item.ETA = fmt.Sprintf("%.1f hours", secondsRemaining/3600)
						} else {
							item.ETA = fmt.Sprintf("%.1f days", secondsRemaining/86400)
						}
					} else {
						item.ETA = "Unknown"
					}
				}

				// Update peer count safely
				item.Peers = len(item.Handle.PeerConns())

				// Update file count if needed
				if item.Handle.Info() != nil {
					item.FileCount = len(item.Handle.Info().Files)
				}

				// Update last update timestamp
				item.LastUpdate = now
			}

			// Use fyne.Do to safely update UI from a goroutine
			fyne.Do(func() {
				// Send notifications for completed downloads
				for hash, completed := range newlyCompleted {
					if completed {
						if item, ok := torrentList[hash]; ok && item != nil {
							a.SendNotification(&fyne.Notification{
								Title:   "Download Complete",
								Content: item.Name,
							})
						}
					}
				}

				// Update status bar with totals
				activeDownloads := 0
				completedDownloads := 0
				var totalDownloadRate int64
				var totalUploadRate int64

				// Calculate counts and rates
				for _, item := range torrentList {
					if item == nil || item.Handle == nil {
						continue
					}

					if item.Progress < 1.0 && item.Status != "Seeding" {
						activeDownloads++
						totalDownloadRate += item.DownloadRate
					} else if item.Progress >= 1.0 {
						completedDownloads++
						totalUploadRate += item.UploadRate
					}
				}

				// Update status bar text
				if statusBar != nil && len(statusBar.Objects) > 0 {
					statusLabel, ok := statusBar.Objects[0].(*widget.Label)
					if ok && statusLabel != nil {
						if activeDownloads > 0 {
							statusLabel.SetText(fmt.Sprintf("Status: Downloading %d torrent(s) at %s, %d completed",
								activeDownloads, HumanReadableRate(totalDownloadRate), completedDownloads))
						} else if len(torrentList) > 0 {
							statusLabel.SetText(fmt.Sprintf("Status: All downloads complete (%d torrents)",
								len(torrentList)))
						} else {
							statusLabel.SetText("Status: Ready")
						}
					}
				}

				// Refresh UI components
				if list != nil {
					list.Refresh()
				}

				// Update details panel
				updateDetailsPanel()
			})

			// Sleep before next update
			time.Sleep(1 * time.Second)
		}
	}()

	// Show the window and run the app
	w.ShowAndRun()
}
