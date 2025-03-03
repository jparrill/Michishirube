package ui

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jparrill/michishirube/internal/data"
	"github.com/jparrill/michishirube/internal/service"
)

// UI represents the main user interface
type UI struct {
	window         fyne.Window
	content        *fyne.Container
	linkService    *service.LinkService
	currentTopicID *int64

	// UI components
	topicTree   *widget.Tree
	linkList    *widget.List
	searchEntry *widget.Entry
	links       []data.Link
	topics      []data.Topic
	topicMap    map[int64]data.Topic
}

// NewUI creates a new UI instance
func NewUI(window fyne.Window, db *sql.DB) *UI {
	linkService := service.NewLinkService(db)

	ui := &UI{
		window:      window,
		linkService: linkService,
		topicMap:    make(map[int64]data.Topic),
	}

	// Initialize UI components
	ui.initUI()

	return ui
}

// Content returns the main UI container
func (ui *UI) Content() fyne.CanvasObject {
	return ui.content
}

// initUI initializes all UI components
func (ui *UI) initUI() {
	// Create search bar
	ui.searchEntry = widget.NewEntry()
	ui.searchEntry.SetPlaceHolder("Search links...")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		ui.searchLinks(ui.searchEntry.Text)
	})
	ui.searchEntry.OnSubmitted = func(s string) {
		ui.searchLinks(s)
	}
	searchContainer := container.NewBorder(nil, nil, nil, searchButton, ui.searchEntry)

	// Create topic tree
	ui.topicTree = ui.createTopicTree()

	// Create topic buttons with labels for clarity
	rootTopicBtn := widget.NewButtonWithIcon("New Root Folder", theme.FolderNewIcon(), func() {
		ui.addRootTopic()
	})
	subTopicBtn := widget.NewButtonWithIcon("New Subfolder", theme.ContentAddIcon(), func() {
		ui.addSubTopic()
	})
	deleteTopicBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), ui.deleteTopic)

	topicContainer := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Topics/Folders"),
			container.NewHBox(rootTopicBtn, subTopicBtn, deleteTopicBtn),
		),
		nil,
		nil, nil,
		container.NewScroll(ui.topicTree),
	)

	// Create link list
	ui.linkList = ui.createLinkList()
	linkContainer := container.NewBorder(
		widget.NewLabel("Links"),
		container.NewHBox(
			widget.NewButtonWithIcon("", theme.ContentAddIcon(), ui.addLink),
		),
		nil, nil,
		container.NewScroll(ui.linkList),
	)

	// Create split view
	split := container.NewHSplit(
		topicContainer,
		linkContainer,
	)
	split.Offset = 0.3 // 30% for topics, 70% for links

	// Main layout
	ui.content = container.NewBorder(
		searchContainer,
		nil, nil, nil,
		split,
	)

	// Load initial data
	ui.refreshTopics()
}

// createTopicTree creates and configures the topic tree widget
func (ui *UI) createTopicTree() *widget.Tree {
	tree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			// Root node
			if id == "" {
				var rootIDs []widget.TreeNodeID
				for _, topic := range ui.topics {
					if !topic.ParentID.Valid {
						rootIDs = append(rootIDs, strconv.FormatInt(topic.ID, 10))
					}
				}
				return rootIDs
			}

			// Child nodes
			var childIDs []widget.TreeNodeID
			parentID, _ := strconv.ParseInt(id, 10, 64)
			for _, topic := range ui.topics {
				if topic.ParentID.Valid && topic.ParentID.Int64 == parentID {
					childIDs = append(childIDs, strconv.FormatInt(topic.ID, 10))
				}
			}
			return childIDs
		},
		func(id widget.TreeNodeID) bool {
			// Check if the node has children
			if id == "" {
				return len(ui.topics) > 0
			}

			parentID, _ := strconv.ParseInt(id, 10, 64)
			for _, topic := range ui.topics {
				if topic.ParentID.Valid && topic.ParentID.Int64 == parentID {
					return true
				}
			}
			return false
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			if id == "" {
				label.SetText("All Topics")
				return
			}

			topicID, _ := strconv.ParseInt(id, 10, 64)
			if topic, ok := ui.topicMap[topicID]; ok {
				label.SetText(topic.Name)
			}
		},
	)

	tree.OnSelected = func(id widget.TreeNodeID) {
		if id == "" {
			ui.currentTopicID = nil
		} else {
			topicID, _ := strconv.ParseInt(id, 10, 64)
			ui.currentTopicID = &topicID
		}
		ui.refreshLinks()
	}

	return tree
}

// createLinkList creates and configures the link list widget
func (ui *UI) createLinkList() *widget.List {
	list := widget.NewList(
		func() int {
			return len(ui.links)
		},
		func() fyne.CanvasObject {
			// Template for link items
			title := widget.NewLabel("Template Title")
			title.TextStyle = fyne.TextStyle{Bold: true}

			url := widget.NewLabel("https://example.com")
			url.TextStyle = fyne.TextStyle{Italic: true}

			textContainer := container.NewVBox(title, url)

			// Placeholder for thumbnail
			thumbnail := canvas.NewImageFromResource(theme.FileImageIcon())
			thumbnail.SetMinSize(fyne.NewSize(32, 32))
			thumbnailContainer := container.NewPadded(thumbnail)

			// Action buttons
			openButton := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), nil)
			moveButton := widget.NewButtonWithIcon("", theme.ContentCutIcon(), nil)
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			actions := container.NewHBox(openButton, moveButton, deleteButton)

			// Create a container with a consistent structure
			// The border container has the following structure:
			// - Top: nil
			// - Bottom: nil
			// - Left: thumbnailContainer
			// - Right: actions
			// - Center: textContainer
			return container.NewBorder(
				nil, nil,
				thumbnailContainer,
				actions,
				textContainer,
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(ui.links) {
				return // Safety check to prevent index out of range
			}

			link := ui.links[id]

			// Safely access the border container
			border, ok := item.(*fyne.Container)
			if !ok {
				log.Printf("Error: item is not a Container")
				return
			}

			// In a border container, the layout is:
			// Objects[0] = Center (textContainer)
			// Objects[1] = Left (thumbnailContainer) - if present
			// Objects[2] = Right (actions) - if present

			// Get the text container (center)
			if len(border.Objects) < 1 {
				log.Printf("Error: border container has no center object")
				return
			}

			textContainer, ok := border.Objects[0].(*fyne.Container)
			if !ok {
				log.Printf("Error: center object is not a Container")
				return
			}

			// Update text labels
			if len(textContainer.Objects) >= 2 {
				if title, ok := textContainer.Objects[0].(*widget.Label); ok {
					title.SetText(link.Name)
				}

				if urlLabel, ok := textContainer.Objects[1].(*widget.Label); ok {
					urlLabel.SetText(link.URL)
				}
			}

			// Get the thumbnail container (left)
			if len(border.Objects) < 2 {
				log.Printf("Error: border container missing left object")
				return
			}

			leftContainer, ok := border.Objects[1].(*fyne.Container)
			if !ok {
				log.Printf("Error: left object is not a Container")
				return
			}

			// Update thumbnail
			if len(leftContainer.Objects) > 0 {
				if thumbnail, ok := leftContainer.Objects[0].(*canvas.Image); ok {
					if link.Thumbnail != "" {
						thumbnail.Resource = theme.FileImageIcon()
					} else {
						thumbnail.Resource = theme.FileIcon()
					}
				}
			}

			// Get the actions container (right)
			if len(border.Objects) < 3 {
				log.Printf("Error: border container missing right object")
				return
			}

			actionsContainer, ok := border.Objects[2].(*fyne.Container)
			if !ok {
				log.Printf("Error: right object is not a Container")
				return
			}

			// Update action buttons
			if len(actionsContainer.Objects) >= 3 {
				// Open button
				if openButton, ok := actionsContainer.Objects[0].(*widget.Button); ok {
					openButton.OnTapped = func() {
						fyne.CurrentApp().OpenURL(mustParseURL(link.URL))
					}
				}

				// Move button
				if moveButton, ok := actionsContainer.Objects[1].(*widget.Button); ok {
					moveButton.OnTapped = func() {
						ui.showMoveDialog(link.ID)
					}
				}

				// Delete button
				if deleteButton, ok := actionsContainer.Objects[2].(*widget.Button); ok {
					deleteButton.OnTapped = func() {
						dialog.ShowConfirm("Delete Link",
							fmt.Sprintf("Are you sure you want to delete '%s'?", link.Name),
							func(confirmed bool) {
								if confirmed {
									err := ui.linkService.DeleteLink(link.ID)
									if err != nil {
										dialog.ShowError(err, ui.window)
										return
									}
									ui.refreshLinks()
								}
							}, ui.window)
					}
				}
			}
		},
	)

	return list
}

// refreshTopics reloads the topic tree from the database
func (ui *UI) refreshTopics() {
	log.Printf("Refreshing topics tree")

	topics, err := ui.linkService.GetTopics(nil)
	if err != nil {
		dialog.ShowError(err, ui.window)
		log.Printf("Error getting topics: %v", err)
		return
	}

	log.Printf("Retrieved %d topics from database", len(topics))

	// Update the topics slice and map
	ui.topics = topics
	ui.topicMap = make(map[int64]data.Topic)

	for _, topic := range topics {
		ui.topicMap[topic.ID] = topic
		log.Printf("Topic: ID=%d, Name=%s, ParentID=%v",
			topic.ID,
			topic.Name,
			topic.ParentID.Int64)
	}

	// Make sure the tree is properly refreshed
	if ui.topicTree != nil {
		ui.topicTree.Refresh()
		log.Printf("Topic tree refreshed")

		// Open the tree to show all topics
		ui.topicTree.OpenAllBranches()
	} else {
		log.Printf("Error: topicTree is nil")
	}
}

// refreshLinks reloads the link list for the current topic
func (ui *UI) refreshLinks() {
	var links []data.Link
	var err error

	if ui.currentTopicID != nil {
		links, err = ui.linkService.GetLinksByTopic(*ui.currentTopicID)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		log.Printf("Refreshed links for topic %d, found %d links", *ui.currentTopicID, len(links))
	} else {
		// If no topic is selected, show all links
		// This is a placeholder - you might want to implement a different behavior
		links = []data.Link{}
		log.Printf("No topic selected, showing empty link list")
	}

	// Update the links slice and refresh the list widget
	ui.links = links

	// Make sure the list is properly refreshed
	if ui.linkList != nil {
		ui.linkList.Refresh()
		log.Printf("Link list refreshed with %d items", len(ui.links))
	} else {
		log.Printf("Error: linkList is nil")
	}
}

// searchLinks searches for links by name
func (ui *UI) searchLinks(searchTerm string) {
	if searchTerm == "" {
		ui.refreshLinks()
		return
	}

	links, err := ui.linkService.SearchLinks(searchTerm)
	if err != nil {
		dialog.ShowError(err, ui.window)
		return
	}

	ui.links = links
	ui.linkList.Refresh()
}

// addTopic shows a dialog to add a new topic
func (ui *UI) addTopic() {
	// Log the current state
	if ui.currentTopicID != nil {
		log.Printf("Adding subtopic to parent topic ID: %d", *ui.currentTopicID)
	} else {
		log.Printf("Adding root-level topic")
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Topic Name")

	dialog.ShowForm("Add Topic", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
		},
		func(confirmed bool) {
			if confirmed {
				name := nameEntry.Text
				if name == "" {
					dialog.ShowError(fmt.Errorf("topic name cannot be empty"), ui.window)
					return
				}

				log.Printf("Creating topic: %s", name)

				// Store the current topic ID locally to avoid any potential changes during async operations
				var parentID *int64
				if ui.currentTopicID != nil {
					copyID := *ui.currentTopicID
					parentID = &copyID
					log.Printf("Using parent ID: %d", *parentID)
				}

				topicID, err := ui.linkService.CreateTopic(name, parentID)
				if err != nil {
					dialog.ShowError(err, ui.window)
					log.Printf("Error creating topic: %v", err)
					return
				}

				log.Printf("Topic created successfully with ID: %d", topicID)

				// Force a refresh of the topics tree
				ui.refreshTopics()

				// Show a confirmation
				dialog.ShowInformation("Success", fmt.Sprintf("Topic '%s' created successfully", name), ui.window)
			}
		}, ui.window)
}

// deleteTopic shows a confirmation dialog to delete the current topic
func (ui *UI) deleteTopic() {
	if ui.currentTopicID == nil {
		dialog.ShowInformation("Error", "Please select a topic to delete", ui.window)
		return
	}

	topic, ok := ui.topicMap[*ui.currentTopicID]
	if !ok {
		dialog.ShowError(fmt.Errorf("topic not found"), ui.window)
		return
	}

	dialog.ShowConfirm("Delete Topic",
		fmt.Sprintf("Are you sure you want to delete '%s' and all its links?", topic.Name),
		func(confirmed bool) {
			if confirmed {
				err := ui.linkService.DeleteTopic(*ui.currentTopicID)
				if err != nil {
					dialog.ShowError(err, ui.window)
					return
				}

				ui.currentTopicID = nil
				ui.refreshTopics()
				ui.refreshLinks()
			}
		}, ui.window)
}

// addLink shows a dialog to add a new link
func (ui *UI) addLink() {
	if ui.currentTopicID == nil {
		dialog.ShowInformation("Error", "Please select a topic first", ui.window)
		return
	}

	topicID := *ui.currentTopicID
	topicName := "Unknown"
	if topic, ok := ui.topicMap[topicID]; ok {
		topicName = topic.Name
	}

	log.Printf("Adding link to topic %s (ID: %d)", topicName, topicID)

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Link Name")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://example.com")

	dialog.ShowForm("Add Link", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
			widget.NewFormItem("URL", urlEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}

			name := nameEntry.Text
			urlStr := urlEntry.Text

			if name == "" || urlStr == "" {
				dialog.ShowError(fmt.Errorf("name and URL cannot be empty"), ui.window)
				return
			}

			log.Printf("Creating link: %s (%s) in topic %d", name, urlStr, topicID)

			// Create the link
			linkID, err := ui.linkService.CreateLink(name, urlStr, topicID)
			if err != nil {
				log.Printf("Error creating link: %v", err)
				dialog.ShowError(err, ui.window)
				return
			}

			log.Printf("Link created with ID: %d", linkID)

			// Force a refresh of the links for the current topic
			ui.refreshLinks()

			// Show a confirmation
			dialog.ShowInformation("Success", fmt.Sprintf("Link '%s' added successfully", name), ui.window)
		}, ui.window)
}

// showMoveDialog shows a dialog to move a link to another topic
func (ui *UI) showMoveDialog(linkID int64) {
	// Create a radio group with all topics
	var options []string
	var optionIDs []int64

	for _, topic := range ui.topics {
		options = append(options, topic.Name)
		optionIDs = append(optionIDs, topic.ID)
	}

	if len(options) == 0 {
		dialog.ShowInformation("Error", "No topics available", ui.window)
		return
	}

	radio := widget.NewRadioGroup(options, nil)
	radio.Selected = options[0]

	dialog.ShowCustomConfirm("Move Link", "Move", "Cancel",
		container.NewVBox(
			widget.NewLabel("Select destination topic:"),
			radio,
		),
		func(confirmed bool) {
			if confirmed {
				selectedIndex := -1
				for i, option := range options {
					if option == radio.Selected {
						selectedIndex = i
						break
					}
				}

				if selectedIndex == -1 {
					dialog.ShowError(fmt.Errorf("no topic selected"), ui.window)
					return
				}

				newTopicID := optionIDs[selectedIndex]
				err := ui.linkService.MoveLink(linkID, newTopicID)
				if err != nil {
					dialog.ShowError(err, ui.window)
					return
				}

				ui.refreshLinks()
			}
		}, ui.window)
}

// addRootTopic creates a new root-level topic (folder)
func (ui *UI) addRootTopic() {
	log.Printf("Adding root-level topic")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Folder Name")

	dialog.ShowForm("Add Root Folder", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}

			name := nameEntry.Text
			if name == "" {
				dialog.ShowError(fmt.Errorf("folder name cannot be empty"), ui.window)
				return
			}

			log.Printf("Creating root folder: %s", name)

			// Root topics have no parent
			topicID, err := ui.linkService.CreateTopic(name, nil)
			if err != nil {
				dialog.ShowError(err, ui.window)
				log.Printf("Error creating root folder: %v", err)
				return
			}

			log.Printf("Root folder created successfully with ID: %d", topicID)

			// Force a refresh of the topics tree
			ui.refreshTopics()

			// Show a confirmation
			dialog.ShowInformation("Success", fmt.Sprintf("Root folder '%s' created successfully", name), ui.window)
		}, ui.window)
}

// addSubTopic creates a new topic (folder) under the currently selected topic
func (ui *UI) addSubTopic() {
	if ui.currentTopicID == nil {
		dialog.ShowInformation("Error", "Please select a parent folder first", ui.window)
		return
	}

	parentID := *ui.currentTopicID
	parentName := "Unknown"
	if topic, ok := ui.topicMap[parentID]; ok {
		parentName = topic.Name
	}

	log.Printf("Adding subfolder to parent folder '%s' (ID: %d)", parentName, parentID)

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Subfolder Name")

	dialog.ShowForm("Add Subfolder", "Add", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Name", nameEntry),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}

			name := nameEntry.Text
			if name == "" {
				dialog.ShowError(fmt.Errorf("subfolder name cannot be empty"), ui.window)
				return
			}

			log.Printf("Creating subfolder: %s under parent %s (ID: %d)", name, parentName, parentID)

			// Store the current topic ID locally to avoid any potential changes during async operations
			copyID := parentID
			parentIDPtr := &copyID

			topicID, err := ui.linkService.CreateTopic(name, parentIDPtr)
			if err != nil {
				dialog.ShowError(err, ui.window)
				log.Printf("Error creating subfolder: %v", err)
				return
			}

			log.Printf("Subfolder created successfully with ID: %d", topicID)

			// Force a refresh of the topics tree
			ui.refreshTopics()

			// Show a confirmation
			dialog.ShowInformation("Success", fmt.Sprintf("Subfolder '%s' created successfully", name), ui.window)
		}, ui.window)
}

// Helper function to safely parse URLs
func mustParseURL(urlStr string) *url.URL {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", urlStr, err)
		return &url.URL{Scheme: "https", Host: "example.com"}
	}
	return parsedURL
}
