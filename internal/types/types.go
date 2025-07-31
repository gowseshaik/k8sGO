package types

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

type ViewType int

const (
	ResourceView ViewType = iota
	DetailView
	HelpView
	ContextView
	DescribeView
	TagsView
	TopView
	ResourcesView
	YamlView
	EventsView
	NamespaceView
	DiagramView
	FeedbackView
	MemoryView
	LogsView
)

type Resource struct {
	Name        string
	Namespace   string
	Type        string
	Status      string
	Ready       string
	Restarts    int
	Age         time.Duration
	CPU         string
	Memory      string
	Description string
	// Additional fields for PVC and other resources
	Volume       string
	Capacity     string
	AccessModes  string
	StorageClass string
}

type FilterOptions struct {
	Status    string
	Namespace string
	Type      string
}

type SortOptions struct {
	Field string
	Order string
}

type PaginatedResult struct {
	Items       []Resource
	CurrentPage int
	TotalPages  int
	TotalItems  int
	HasNext     bool
	HasPrev     bool
}

type AppState struct {
	CurrentView      ViewType
	SelectedResource Resource
	SearchQuery      string
	FilterCriteria   FilterOptions
	SortOrder        SortOptions

	ToolName    string
	Version     string
	ShowBanner  bool
	BannerStyle BannerStyle

	CurrentPage   int
	PageSize      int
	TotalItems    int
	TotalPages    int
	SelectedIndex int

	FocusedElement string
	ScreenReader   bool
	HighContrast   bool

	Resources    []Resource
	ResourceType string
	LastUpdate   time.Time

	Context   string
	Namespace string
	Tool      string
	User      string

	// Describe view data
	DescribeOutput string

	// Navigation history for ESC/Tab and backspace
	PreviousView     ViewType
	PreviousResource string

	// Scroll position for describe view
	DescribeScrollOffset int

	// Tags view data
	TagsOutput       string
	TagsScrollOffset int

	// Top view data
	TopOutput       string
	TopScrollOffset int

	// Resources view data
	ResourcesOutput       string
	ResourcesScrollOffset int

	// YAML view data
	YamlOutput       string
	YamlScrollOffset int

	// Events view data
	EventsOutput       string
	EventsScrollOffset int

	// Diagram view data
	DiagramOutput       string
	DiagramScrollOffset int

	// Memory view data
	MemoryOutput       string
	MemoryScrollOffset int

	// Logs view data
	LogsOutput       string
	LogsScrollOffset int

	// Status message for user feedback
	StatusMessage string

	// Text selection state
	SelectionActive bool
	SelectionStartX int
	SelectionStartY int
	SelectionEndX   int
	SelectionEndY   int
	SelectedText    string

	// Feedback form state
	FeedbackText       string
	FeedbackCursorPos  int
	ShowFeedbackForm   bool
	FeedbackSubmitting bool

	// Manual sorting state
	SortField     string // "name", "age", "status", etc.
	SortDirection string // "asc" or "desc"
}

type BannerStyle struct {
	Icon        string
	Colors      lipgloss.Style
	Compact     bool
	ShowVersion bool
}

type BannerManager struct {
	ToolName  string
	Version   string
	BuildDate string
	GitCommit string
	Enabled   bool
	Style     BannerStyle
}

type ResourcePaginator struct {
	PageSize    int
	CurrentPage int
	TotalItems  int
}

type ClusterClient interface {
	GetPods(namespace string) ([]Resource, error)
	GetServices(namespace string) ([]Resource, error)
	GetDeployments(namespace string) ([]Resource, error)
	SwitchContext(context string) error
	GetContexts() ([]string, error)
	GetCurrentContext() (string, error)
	GetNamespaces() ([]string, error)
}
