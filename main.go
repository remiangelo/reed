package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/anacrolix/torrent"
)

// ReedTheme is a modern, minimalist theme with light and dark mode support
type ReedTheme struct {
	fyne.Theme
	isDark bool
}

// Color returns the color for the specified name and theme
func (t *ReedTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Purple color scheme based on the Reed logo
	primaryColor := color.NRGBA{R: 108, G: 92, B: 231, A: 255}       // #6c5ce7 from logo
	lightPrimaryColor := color.NRGBA{R: 162, G: 155, B: 254, A: 255} // #a29bfe from logo

	if t.isDark {
		// Dark theme colors
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 30, G: 30, B: 40, A: 255} // Dark background
		case theme.ColorNameButton:
			return primaryColor
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 60, G: 60, B: 70, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 240, G: 240, B: 250, A: 255} // Light text
		case theme.ColorNameHover:
			return lightPrimaryColor
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 45, G: 45, B: 55, A: 255}
		case theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 150, G: 150, B: 160, A: 255}
		case theme.ColorNamePressed:
			return color.NRGBA{R: 90, G: 80, B: 200, A: 255}
		case theme.ColorNamePrimary:
			return primaryColor
		case theme.ColorNameScrollBar:
			return color.NRGBA{R: 60, G: 60, B: 70, A: 255}
		case theme.ColorNameShadow:
			return color.NRGBA{R: 0, G: 0, B: 0, A: 80}
		default:
			return t.Theme.Color(name, variant)
		}
	} else {
		// Light theme colors
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 250, G: 250, B: 255, A: 255} // Light background
		case theme.ColorNameButton:
			return primaryColor
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 200, G: 200, B: 210, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 30, G: 30, B: 40, A: 255} // Dark text
		case theme.ColorNameHover:
			return lightPrimaryColor
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 150, G: 150, B: 160, A: 255}
		case theme.ColorNamePressed:
			return color.NRGBA{R: 90, G: 80, B: 200, A: 255}
		case theme.ColorNamePrimary:
			return primaryColor
		case theme.ColorNameScrollBar:
			return color.NRGBA{R: 220, G: 220, B: 230, A: 255}
		case theme.ColorNameShadow:
			return color.NRGBA{R: 0, G: 0, B: 0, A: 40}
		default:
			return t.Theme.Color(name, variant)
		}
	}
}

// ToggleDarkMode switches between light and dark mode
func (t *ReedTheme) ToggleDarkMode() {
	t.isDark = !t.isDark
}

// Icon returns the appropriate icon for the iconName and variant
func (t *ReedTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.Theme.Icon(name)
}

// Font returns the font resource for the specified style and size
func (t *ReedTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.Theme.Font(style)
}

// Size returns the size for the specified name
func (t *ReedTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6 // Slightly more padding for a cleaner look
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 8 // Thinner scrollbar
	case theme.SizeNameScrollBarSmall:
		return 4 // Even thinner small scrollbar
	case theme.SizeNameText:
		return 13 // Slightly larger text for better readability
	case theme.SizeNameHeadingText:
		return 18 // Larger heading text
	case theme.SizeNameSubHeadingText:
		return 15 // Larger subheading text
	case theme.SizeNameCaptionText:
		return 11 // Larger caption text
	case theme.SizeNameSeparatorThickness:
		return 1 // Thinner separators for a cleaner look
	default:
		return t.Theme.Size(name)
	}
}

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

	// Create our custom theme with light mode as default
	appTheme := &ReedTheme{Theme: theme.DefaultTheme(), isDark: false}
	a.Settings().SetTheme(appTheme)

	w := a.NewWindow("Reed Torrent Client")
	w.Resize(fyne.NewSize(1024, 768)) // Larger default size like Vuze

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

	// Enhanced torrent list widget with Vuze-like styling
	list := widget.NewList(
		func() int {
			return len(torrentList)
		},
		func() fyne.CanvasObject {
			// Create a more modern, minimalist template
			nameText := canvas.NewText("Torrent Name", color.NRGBA{R: 108, G: 92, B: 231, A: 255}) // Purple from logo
			nameText.TextSize = 15
			nameText.TextStyle = fyne.TextStyle{Bold: true}

			progressBar := widget.NewProgressBar()
			progressBar.Min = 0
			progressBar.Max = 1

			// Create a container with all the torrent information in a cleaner layout
			return container.NewVBox(
				container.NewPadded(
					container.NewHBox(
						widget.NewIcon(theme.FileIcon()),
						nameText,
					),
				),
				progressBar,
				container.NewPadded(
					container.NewGridWithColumns(4,
						container.NewVBox(
							widget.NewLabelWithStyle("Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabel("Status"),
						),
						container.NewVBox(
							widget.NewLabelWithStyle("Size", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabel("Size"),
						),
						container.NewVBox(
							widget.NewLabelWithStyle("Speed", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabel("Speed"),
						),
						container.NewVBox(
							widget.NewLabelWithStyle("ETA", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
							widget.NewLabel("ETA"),
						),
					),
				),
				widget.NewSeparator(),
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
			if !ok || len(vbox.Objects) < 4 {
				return
			}

			// Top row with icon and name (now inside a padded container)
			paddedHBox, ok := vbox.Objects[0].(*fyne.Container)
			if !ok || len(paddedHBox.Objects) < 1 {
				return
			}

			hbox, ok := paddedHBox.Objects[0].(*fyne.Container)
			if !ok || len(hbox.Objects) < 2 {
				return
			}

			nameText, ok := hbox.Objects[1].(*canvas.Text)
			if !ok {
				return
			}

			// Progress bar
			progressBar, ok := vbox.Objects[1].(*widget.ProgressBar)
			if !ok {
				return
			}

			// Grid with stats (now inside a padded container)
			paddedStatsGrid, ok := vbox.Objects[2].(*fyne.Container)
			if !ok || len(paddedStatsGrid.Objects) < 1 {
				return
			}

			statsGrid, ok := paddedStatsGrid.Objects[0].(*fyne.Container)
			if !ok || len(statsGrid.Objects) < 4 {
				return
			}

			// Status column
			statusBox, ok := statsGrid.Objects[0].(*fyne.Container)
			if !ok || len(statusBox.Objects) < 2 {
				return
			}
			statusLabel, ok := statusBox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			// Size column
			sizeBox, ok := statsGrid.Objects[1].(*fyne.Container)
			if !ok || len(sizeBox.Objects) < 2 {
				return
			}
			sizeLabel, ok := sizeBox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			// Speed column
			speedBox, ok := statsGrid.Objects[2].(*fyne.Container)
			if !ok || len(speedBox.Objects) < 2 {
				return
			}
			speedLabel, ok := speedBox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			// ETA column
			etaBox, ok := statsGrid.Objects[3].(*fyne.Container)
			if !ok || len(etaBox.Objects) < 2 {
				return
			}
			etaLabel, ok := etaBox.Objects[1].(*widget.Label)
			if !ok {
				return
			}

			// Set values safely
			nameText.Text = torrentItem.Name
			nameText.Refresh()

			progressBar.Value = torrentItem.Progress

			// Set status with color based on state
			statusLabel.SetText(torrentItem.Status)

			// Set size
			sizeLabel.SetText(HumanReadableSize(torrentItem.Size))

			// Set speed
			if torrentItem.DownloadRate > 0 {
				speedLabel.SetText("↓ " + HumanReadableRate(torrentItem.DownloadRate))
			} else if torrentItem.UploadRate > 0 {
				speedLabel.SetText("↑ " + HumanReadableRate(torrentItem.UploadRate))
			} else {
				speedLabel.SetText("-")
			}

			// Set ETA
			etaLabel.SetText(torrentItem.ETA)
		},
	)

	// Set up list selection
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
	}

	// Enhanced status bar with more detailed information (like Vuze)
	downloadSpeedLabel := widget.NewLabel("↓ 0 B/s")
	uploadSpeedLabel := widget.NewLabel("↑ 0 B/s")
	activeTorrentsLabel := widget.NewLabel("0 Active")
	completedTorrentsLabel := widget.NewLabel("0 Complete")

	// Create a more modern, minimalist status bar
	statusText := widget.NewLabel("Ready")
	statusText.Alignment = fyne.TextAlignLeading

	// Create a container for the status indicators with some padding
	statusIndicators := container.NewPadded(
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			statusText,
			widget.NewSeparator(),
			widget.NewIcon(theme.DownloadIcon()),
			downloadSpeedLabel,
			widget.NewSeparator(),
			widget.NewIcon(theme.UploadIcon()),
			uploadSpeedLabel,
			widget.NewSeparator(),
			activeTorrentsLabel,
			widget.NewSeparator(),
			completedTorrentsLabel,
		),
	)

	// Create a container for the directory info with right alignment
	dirInfo := container.NewPadded(
		container.NewHBox(
			layout.NewSpacer(),
			widget.NewIcon(theme.FolderIcon()),
			widget.NewLabel(cfg.DataDir),
		),
	)

	// Combine the status indicators and directory info in a border layout
	statusBar := container.NewBorder(
		nil, nil, statusIndicators, dirInfo,
		nil,
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
				defer func(reader fyne.URIReadCloser) {
					err := reader.Close()
					if err != nil {

					}
				}(reader)

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
		widget.NewToolbarAction(theme.ColorPaletteIcon(), func() {
			// Toggle between light and dark mode
			appTheme.ToggleDarkMode()
			a.Settings().SetTheme(appTheme)
		}),
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

	// Function to update the details panel with a tabbed interface like Vuze
	updateDetailsPanel = func() {
		// Clear the container
		detailsContainer.Objects = nil

		if selectedIndex < 0 {
			// Create a styled "no selection" message
			noSelectionText := canvas.NewText("No torrent selected", color.NRGBA{R: 100, G: 100, B: 100, A: 255})
			noSelectionText.TextSize = 16
			noSelectionText.Alignment = fyne.TextAlignCenter

			detailsContainer.Add(container.NewCenter(noSelectionText))
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
			errorText := canvas.NewText("Torrent not found or none selected", color.NRGBA{R: 200, G: 50, B: 50, A: 255})
			errorText.TextSize = 16
			errorText.Alignment = fyne.TextAlignCenter

			detailsContainer.Add(container.NewCenter(errorText))
			detailsContainer.Refresh()
			return
		}

		// Additional safety check
		if selectedTorrent.Handle == nil || selectedTorrent.Handle.Info() == nil {
			loadingText := canvas.NewText("Loading torrent information...", color.NRGBA{R: 100, G: 100, B: 100, A: 255})
			loadingText.TextSize = 16
			loadingText.Alignment = fyne.TextAlignCenter

			detailsContainer.Add(container.NewCenter(loadingText))
			detailsContainer.Refresh()
			return
		}

		// Add torrent title with styled text
		titleText := canvas.NewText(selectedTorrent.Name, color.NRGBA{R: 66, G: 133, B: 244, A: 255})
		titleText.TextSize = 18
		titleText.TextStyle = fyne.TextStyle{Bold: true}

		detailsContainer.Add(titleText)
		detailsContainer.Add(widget.NewSeparator())

		// Create action buttons with icons
		actionsContainer := container.NewHBox(
			widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
				dialog.ShowInformation("Not Implemented", "Start functionality will be added soon.", w)
			}),
			widget.NewButtonWithIcon("Pause", theme.MediaPauseIcon(), func() {
				dialog.ShowInformation("Not Implemented", "Pause functionality will be added soon.", w)
			}),
			widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
				dialog.ShowInformation("Not Implemented", "Stop functionality will be added soon.", w)
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Open Folder", theme.FolderOpenIcon(), func() {
				dialog.ShowInformation("Open Folder", "This will open the folder containing the downloaded files.", w)
			}),
		)
		detailsContainer.Add(actionsContainer)
		detailsContainer.Add(widget.NewSeparator())

		// Create tabs for different types of information (like Vuze)

		// General tab
		generalTab := container.NewVBox()

		// Create a more detailed info form with styled labels
		infoGrid := container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabelWithStyle("Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(selectedTorrent.Status),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Size:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(HumanReadableSize(selectedTorrent.Size)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Downloaded:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(HumanReadableSize(selectedTorrent.Downloaded)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Progress:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("%.1f%%", selectedTorrent.Progress*100)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Download Speed:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(HumanReadableRate(selectedTorrent.DownloadRate)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Upload Speed:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(HumanReadableRate(selectedTorrent.UploadRate)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Peers:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("%d", selectedTorrent.Peers)),
			),
			container.NewVBox(
				widget.NewLabelWithStyle("Seeds:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("%d", selectedTorrent.Seeds)),
			),
		)

		// Add ETA if downloading
		if selectedTorrent.Progress < 1.0 && selectedTorrent.DownloadRate > 0 {
			etaBox := container.NewVBox(
				widget.NewLabelWithStyle("ETA:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(selectedTorrent.ETA),
			)
			infoGrid.Add(etaBox)
		}

		// Add metadata info
		addedBox := container.NewVBox(
			widget.NewLabelWithStyle("Added:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(selectedTorrent.AddedAt.Format("2006-01-02 15:04:05")),
		)
		infoGrid.Add(addedBox)

		generalTab.Add(infoGrid)

		// Files tab
		filesTab := container.NewVBox()

		// If the torrent has files, show them
		if selectedTorrent != nil && selectedTorrent.Handle != nil &&
			selectedTorrent.Handle.Info() != nil {

			// Create a header for the files list
			filesHeader := container.NewHBox(
				widget.NewLabelWithStyle("Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				layout.NewSpacer(),
				widget.NewLabelWithStyle("Progress", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle("Size", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
			)
			filesTab.Add(filesHeader)
			filesTab.Add(widget.NewSeparator())

			if len(selectedTorrent.Handle.Info().Files) > 0 {
				// Multiple files
				filesList := widget.NewList(
					func() int {
						// Double-check that info is still available
						if selectedTorrent != nil && selectedTorrent.Handle != nil &&
							selectedTorrent.Handle.Info() != nil {
							return len(selectedTorrent.Handle.Info().Files)
						}
						return 0
					},
					func() fyne.CanvasObject {
						return container.NewBorder(
							nil, nil,
							container.NewHBox(widget.NewIcon(theme.FileIcon())),
							container.NewHBox(
								widget.NewProgressBar(),
								widget.NewLabel("Size"),
							),
							widget.NewLabel("Filename"),
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

						border := obj.(*fyne.Container)
						filenameLabel := border.Objects[0].(*widget.Label)
						rightContainer := border.Objects[1].(*fyne.Container)
						progressBar := rightContainer.Objects[0].(*widget.ProgressBar)
						sizeLabel := rightContainer.Objects[1].(*widget.Label)

						// Get the filename from the path
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

				// Wrap the files list in a scroll container
				filesScroll := container.NewVScroll(filesList)
				filesScroll.SetMinSize(fyne.NewSize(0, 200))
				filesTab.Add(filesScroll)
			} else {
				// Single file torrent
				filesTab.Add(widget.NewLabel(selectedTorrent.Name))
				filesTab.Add(container.NewHBox(
					widget.NewProgressBar(),
					widget.NewLabel(HumanReadableSize(selectedTorrent.Size)),
				))
			}
		} else {
			filesTab.Add(widget.NewLabel("No file information available"))
		}

		// Peers tab (placeholder for now)
		peersTab := container.NewVBox(
			widget.NewLabelWithStyle("Peers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabel(fmt.Sprintf("Connected to %d peers", selectedTorrent.Peers)),
			widget.NewLabel("Detailed peer information will be implemented in a future update."),
		)

		// Create the tab container for details
		detailsTabs := container.NewAppTabs(
			container.NewTabItem("General", generalTab),
			container.NewTabItem("Files", filesTab),
			container.NewTabItem("Peers", peersTab),
		)

		detailsContainer.Add(detailsTabs)
		detailsContainer.Refresh()
	}

	// Set up list selection to update the details panel - this overrides the previous OnSelected
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
		updateDetailsPanel()
	}

	// Create a header with app logo and title
	var header *fyne.Container

	// Load the SVG logo
	logoResource, err := fyne.LoadResourceFromPath("icon.svg")
	if err != nil {
		log.Printf("Error loading logo: %v", err)
		// Fallback to text logo if SVG loading fails
		headerLogo := canvas.NewText("REED", color.NRGBA{R: 108, G: 92, B: 231, A: 255}) // Purple from logo
		headerLogo.TextSize = 24
		headerLogo.TextStyle = fyne.TextStyle{Bold: true}

		headerTitle := canvas.NewText("Torrent Client", color.NRGBA{R: 100, G: 100, B: 100, A: 255})
		headerTitle.TextSize = 18

		header = container.NewHBox(
			headerLogo,
			widget.NewLabel(" "),
			headerTitle,
			layout.NewSpacer(),
		)
	} else {
		// Create an image with the SVG logo
		logoImage := canvas.NewImageFromResource(logoResource)
		logoImage.SetMinSize(fyne.NewSize(120, 36)) // Set appropriate size for the logo
		logoImage.FillMode = canvas.ImageFillContain

		header = container.NewHBox(
			logoImage,
			layout.NewSpacer(),
		)
	}

	// Create a tabbed interface for different views (like Vuze)
	// Library tab - contains the torrent list and details
	libraryTab := container.NewBorder(
		nil, nil, nil, nil,
		container.NewHSplit(
			container.NewVBox(
				// Enhanced torrent list with category header
				container.NewVBox(
					container.NewHBox(
						widget.NewLabelWithStyle("Library", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						layout.NewSpacer(),
						widget.NewLabel(fmt.Sprintf("%d Torrents", len(torrentList))),
					),
					widget.NewSeparator(),
					container.NewVBox(list),
				),
			),
			container.NewScroll(detailsContainer),
		),
	)

	// Files tab - will show all files across torrents
	filesTab := container.NewVBox(
		widget.NewLabelWithStyle("All Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Files view will be implemented in a future update."),
	)

	// Stats tab - will show statistics
	statsTab := container.NewVBox(
		widget.NewLabelWithStyle("Statistics", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Statistics view will be implemented in a future update."),
	)

	// Create the tab container
	mainTabs := container.NewAppTabs(
		container.NewTabItem("Library", libraryTab),
		container.NewTabItem("Files", filesTab),
		container.NewTabItem("Statistics", statsTab),
	)
	mainTabs.SetTabLocation(container.TabLocationTop)

	// Create the main layout with the toolbar at the top
	content := container.NewBorder(
		container.NewVBox(
			header,
			toolbar,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			statusBar,
		),
		nil,
		nil,
		mainTabs,
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

				// Update the status text directly using the variables we created earlier
				if activeDownloads > 0 {
					statusText.SetText("Downloading")
				} else if len(torrentList) > 0 {
					statusText.SetText("Idle")
				} else {
					statusText.SetText("Ready")
				}

				// Update the speed labels
				downloadSpeedLabel.SetText("↓ " + HumanReadableRate(totalDownloadRate))
				uploadSpeedLabel.SetText("↑ " + HumanReadableRate(totalUploadRate))

				// Update the torrent count labels
				activeTorrentsLabel.SetText(fmt.Sprintf("%d Active", activeDownloads))
				completedTorrentsLabel.SetText(fmt.Sprintf("%d Complete", completedDownloads))

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
