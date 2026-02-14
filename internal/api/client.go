package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	clerrors "github.com/buddyh/todoist-cli/internal/errors"
)

const (
	BaseURL    = "https://api.todoist.com/api/v1"
	maxRetries = 3
)

// paginatedResponse wraps list endpoints that return cursor-paginated results.
type paginatedResponse struct {
	Results    json.RawMessage `json:"results"`
	NextCursor *string        `json:"next_cursor"`
}

// unmarshalList parses a paginated API response, extracting the results array
// into the target slice. Falls back to direct unmarshal for non-paginated responses.
func unmarshalList(data []byte, v interface{}) error {
	var page paginatedResponse
	if err := json.Unmarshal(data, &page); err == nil && page.Results != nil {
		return json.Unmarshal(page.Results, v)
	}
	return json.Unmarshal(data, v)
}

// Client is a Todoist API client
type Client struct {
	token      string
	httpClient *http.Client
	debug      bool
}

// NewClient creates a new Todoist API client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetDebug enables HTTP request/response tracing to stderr.
func (c *Client) SetDebug(enabled bool) {
	c.debug = enabled
}

// request makes an authenticated request to the Todoist API
func (c *Client) request(method, endpoint string, data interface{}) ([]byte, error) {
	return c.requestCtx(context.Background(), method, endpoint, data)
}

// requestCtx makes an authenticated request with context support and retry logic.
func (c *Client) requestCtx(ctx context.Context, method, endpoint string, data interface{}) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s", BaseURL, endpoint)

	var bodyBytes []byte
	if data != nil {
		if method == "GET" {
			// For GET, encode as query params
			if params, ok := data.(map[string]string); ok {
				values := url.Values{}
				for k, v := range params {
					if v != "" {
						values.Set(k, v)
					}
				}
				if len(values) > 0 {
					reqURL = fmt.Sprintf("%s?%s", reqURL, values.Encode())
				}
			}
		} else {
			// For POST/DELETE, encode as JSON body
			var err error
			bodyBytes, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request: %w", err)
			}
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			if lastErr != nil {
				if waitErr, ok := lastErr.(*retryAfterError); ok {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(waitErr.after):
					}
				} else {
					backoff := time.Duration(1<<uint(attempt-1)) * time.Second
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(backoff):
					}
				}
			}
		}

		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
		req.Header.Set("Content-Type", "application/json")

		start := time.Now()
		if c.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, reqURL)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if c.debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] error: %v (%s)\n", err, time.Since(start))
			}
			return nil, clerrors.WrapNetworkError("request failed", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, clerrors.WrapNetworkError("failed to read response", err)
		}

		if c.debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] %d %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
		}

		if resp.StatusCode == 429 && attempt < maxRetries {
			wait := 5 * time.Second
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			lastErr = &retryAfterError{after: wait}
			if c.debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] rate limited, retrying in %s\n", wait)
			}
			continue
		}

		if resp.StatusCode >= 400 {
			apiErr := fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
			switch resp.StatusCode {
			case 401, 403:
				return nil, clerrors.WrapAuthError("authentication failed", apiErr)
			default:
				return nil, apiErr
			}
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

type retryAfterError struct {
	after time.Duration
}

func (e *retryAfterError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.after)
}

// =============================================================================
// TASKS
// =============================================================================

// Task represents a Todoist task
type Task struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Description string   `json:"description"`
	ProjectID   string   `json:"project_id"`
	SectionID   string   `json:"section_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	Order       int      `json:"order"`
	Priority    int      `json:"priority"`
	Due         *Due     `json:"due,omitempty"`
	URL         string   `json:"url"`
	Labels      []string `json:"labels"`
	CreatedAt   string   `json:"created_at"`
	CreatorID   string   `json:"creator_id"`
	Assignee    string   `json:"assignee_id,omitempty"`
	Assigner    string   `json:"assigner_id,omitempty"`
	IsCompleted bool     `json:"is_completed"`
}

// Due represents a task due date
type Due struct {
	Date        string `json:"date"`
	String      string `json:"string"`
	Datetime    string `json:"datetime,omitempty"`
	IsRecurring bool   `json:"is_recurring"`
	Timezone    string `json:"timezone,omitempty"`
}

// GetTasks returns all active tasks with optional filters
func (c *Client) GetTasks(projectID, filter string) ([]Task, error) {
	params := map[string]string{}
	if projectID != "" {
		params["project_id"] = projectID
	}
	if filter != "" {
		params["filter"] = filter
	}

	var data interface{}
	if len(params) > 0 {
		data = params
	}

	resp, err := c.request("GET", "tasks", data)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := unmarshalList(resp, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	return tasks, nil
}

// GetTask returns a single task by ID
func (c *Client) GetTask(taskID string) (*Task, error) {
	resp, err := c.request("GET", fmt.Sprintf("tasks/%s", taskID), nil)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return &task, nil
}

// AddTaskParams contains parameters for creating a task
type AddTaskParams struct {
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	SectionID   string   `json:"section_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	AssigneeID  string   `json:"assignee_id,omitempty"`
}

// AddTask creates a new task
func (c *Client) AddTask(params AddTaskParams) (*Task, error) {
	resp, err := c.request("POST", "tasks", params)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return &task, nil
}

// UpdateTaskParams contains parameters for updating a task
type UpdateTaskParams struct {
	Content     string   `json:"content,omitempty"`
	Description string   `json:"description,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	AssigneeID  string   `json:"assignee_id,omitempty"`
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(taskID string, params UpdateTaskParams) (*Task, error) {
	resp, err := c.request("POST", fmt.Sprintf("tasks/%s", taskID), params)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return &task, nil
}

// CompleteTask marks a task as complete
func (c *Client) CompleteTask(taskID string) error {
	_, err := c.request("POST", fmt.Sprintf("tasks/%s/close", taskID), nil)
	return err
}

// ReopenTask reopens a completed task
func (c *Client) ReopenTask(taskID string) error {
	_, err := c.request("POST", fmt.Sprintf("tasks/%s/reopen", taskID), nil)
	return err
}

// DeleteTask permanently deletes a task
func (c *Client) DeleteTask(taskID string) error {
	_, err := c.request("DELETE", fmt.Sprintf("tasks/%s", taskID), nil)
	return err
}

// ReorderTask sets the order of a task using the Sync API
func (c *Client) ReorderTask(taskID string, order int) error {
	uuid := fmt.Sprintf("%d", time.Now().UnixNano())

	commands := []map[string]interface{}{
		{
			"type": "item_reorder",
			"uuid": uuid,
			"args": map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": taskID, "child_order": order},
				},
			},
		},
	}

	params := map[string]interface{}{
		"commands": commands,
	}

	_, err := c.request("POST", "sync", params)
	return err
}

// =============================================================================
// PROJECTS
// =============================================================================

// Project represents a Todoist project
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	ParentID       string `json:"parent_id,omitempty"`
	Order          int    `json:"order"`
	CommentCount   int    `json:"comment_count"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox"`
	ViewStyle      string `json:"view_style"`
	URL            string `json:"url"`
}

// GetProjects returns all projects
func (c *Client) GetProjects() ([]Project, error) {
	resp, err := c.request("GET", "projects", nil)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := unmarshalList(resp, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}

	return projects, nil
}

// GetProject returns a single project by ID
func (c *Client) GetProject(projectID string) (*Project, error) {
	resp, err := c.request("GET", fmt.Sprintf("projects/%s", projectID), nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	return &project, nil
}

// FindProject finds a project by name (case-insensitive partial match)
func (c *Client) FindProject(name string) (*Project, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), nameLower) {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", name)
}

// AddProjectParams contains parameters for creating a project
type AddProjectParams struct {
	Name       string `json:"name"`
	Color      string `json:"color,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
}

// AddProject creates a new project
func (c *Client) AddProject(params AddProjectParams) (*Project, error) {
	resp, err := c.request("POST", "projects", params)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}

	return &project, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(projectID string) error {
	_, err := c.request("DELETE", fmt.Sprintf("projects/%s", projectID), nil)
	return err
}

// =============================================================================
// SECTIONS
// =============================================================================

// Section represents a project section
type Section struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Order     int    `json:"order"`
	Name      string `json:"name"`
}

// GetSections returns all sections, optionally filtered by project
func (c *Client) GetSections(projectID string) ([]Section, error) {
	var data interface{}
	if projectID != "" {
		data = map[string]string{"project_id": projectID}
	}

	resp, err := c.request("GET", "sections", data)
	if err != nil {
		return nil, err
	}

	var sections []Section
	if err := unmarshalList(resp, &sections); err != nil {
		return nil, fmt.Errorf("failed to parse sections: %w", err)
	}

	return sections, nil
}

// AddSection creates a new section
func (c *Client) AddSection(name, projectID string) (*Section, error) {
	params := map[string]string{
		"name":       name,
		"project_id": projectID,
	}

	resp, err := c.request("POST", "sections", params)
	if err != nil {
		return nil, err
	}

	var section Section
	if err := json.Unmarshal(resp, &section); err != nil {
		return nil, fmt.Errorf("failed to parse section: %w", err)
	}

	return &section, nil
}

// =============================================================================
// LABELS
// =============================================================================

// Label represents a Todoist label
type Label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Order      int    `json:"order"`
	IsFavorite bool   `json:"is_favorite"`
}

// GetLabels returns all labels
func (c *Client) GetLabels() ([]Label, error) {
	resp, err := c.request("GET", "labels", nil)
	if err != nil {
		return nil, err
	}

	var labels []Label
	if err := unmarshalList(resp, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}

	return labels, nil
}

// AddLabel creates a new label
func (c *Client) AddLabel(name, color string) (*Label, error) {
	params := map[string]string{"name": name}
	if color != "" {
		params["color"] = color
	}

	resp, err := c.request("POST", "labels", params)
	if err != nil {
		return nil, err
	}

	var label Label
	if err := json.Unmarshal(resp, &label); err != nil {
		return nil, fmt.Errorf("failed to parse label: %w", err)
	}

	return &label, nil
}

// =============================================================================
// COMMENTS
// =============================================================================

// Comment represents a Todoist comment
type Comment struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Content   string `json:"content"`
	PostedAt  string `json:"posted_at"`
}

// GetComments returns comments for a task or project
func (c *Client) GetComments(taskID, projectID string) ([]Comment, error) {
	return c.GetCommentsCtx(context.Background(), taskID, projectID)
}

// GetCommentsCtx returns comments with context support
func (c *Client) GetCommentsCtx(ctx context.Context, taskID, projectID string) ([]Comment, error) {
	params := map[string]string{}
	if taskID != "" {
		params["task_id"] = taskID
	} else if projectID != "" {
		params["project_id"] = projectID
	}

	resp, err := c.requestCtx(ctx, "GET", "comments", params)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := unmarshalList(resp, &comments); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}

	return comments, nil
}

// AddComment adds a comment to a task or project
func (c *Client) AddComment(content, taskID, projectID string) (*Comment, error) {
	params := map[string]string{"content": content}
	if taskID != "" {
		params["task_id"] = taskID
	} else if projectID != "" {
		params["project_id"] = projectID
	}

	resp, err := c.request("POST", "comments", params)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := json.Unmarshal(resp, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment: %w", err)
	}

	return &comment, nil
}

// =============================================================================
// COLLABORATORS
// =============================================================================

// Collaborator represents a project collaborator
type Collaborator struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetCollaborators returns collaborators for a project
func (c *Client) GetCollaborators(projectID string) ([]Collaborator, error) {
	resp, err := c.request("GET", fmt.Sprintf("projects/%s/collaborators", projectID), nil)
	if err != nil {
		return nil, err
	}

	var collaborators []Collaborator
	if err := unmarshalList(resp, &collaborators); err != nil {
		return nil, fmt.Errorf("failed to parse collaborators: %w", err)
	}

	return collaborators, nil
}

// =============================================================================
// COMPLETED TASKS (Sync API)
// =============================================================================

// CompletedTask represents a completed task from the Sync API
type CompletedTask struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id"`
	Content     string `json:"content"`
	ProjectID   string `json:"project_id"`
	CompletedAt string `json:"completed_at"`
}

// CompletedTasksResponse is the response from the completed tasks endpoint
type CompletedTasksResponse struct {
	Items []CompletedTask `json:"items"`
}

// MoveTask moves a task to a different section or project using the Sync API
func (c *Client) MoveTask(taskID, sectionID, projectID string) error {
	// Generate a UUID for the command
	uuid := fmt.Sprintf("%d", time.Now().UnixNano())

	args := map[string]string{"id": taskID}
	if sectionID != "" {
		args["section_id"] = sectionID
	} else if projectID != "" {
		args["project_id"] = projectID
	}

	commands := []map[string]interface{}{
		{
			"type": "item_move",
			"uuid": uuid,
			"args": args,
		},
	}

	params := map[string]interface{}{
		"commands": commands,
	}

	_, err := c.request("POST", "sync", params)
	return err
}

// GetCompletedTasks returns completed tasks
func (c *Client) GetCompletedTasks(projectID, since, until string, limit int) (*CompletedTasksResponse, error) {
	params := map[string]interface{}{
		"limit": limit,
	}
	if projectID != "" {
		params["project_id"] = projectID
	}
	if since != "" {
		params["since"] = since
	}
	if until != "" {
		params["until"] = until
	}

	resp, err := c.request("POST", "completed/get_all", params)
	if err != nil {
		return nil, err
	}

	var result CompletedTasksResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse completed tasks: %w", err)
	}

	return &result, nil
}
