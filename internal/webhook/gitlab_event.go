package webhook

// No external imports needed - GitLabTime is in same package

// GitLabWebhookEvent represents the base GitLab webhook event
type GitLabWebhookEvent struct {
	ObjectKind string `json:"object_kind"` // "merge_request"
	EventType  string `json:"event_type"`  // e.g., "merge_request"
}

// GitLabMergeRequestEvent represents GitLab merge request webhook payload
type GitLabMergeRequestEvent struct {
	ObjectKind       string                       `json:"object_kind"`
	EventType        string                       `json:"event_type"`
	User             GitLabUser                   `json:"user"`
	Project          GitLabProject                `json:"project"`
	ObjectAttributes GitLabMergeRequestAttributes `json:"object_attributes"`
	Labels           []GitLabLabel                `json:"labels"`
	Changes          GitLabMRChanges              `json:"changes"`
	Repository       GitLabRepository             `json:"repository"`
}

// GitLabUser represents a GitLab user
type GitLabUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// GitLabProject represents a GitLab project
type GitLabProject struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	AvatarURL         string `json:"avatar_url"`
	GitSSHURL         string `json:"git_ssh_url"`
	GitHTTPURL        string `json:"git_http_url"`
	Namespace         string `json:"namespace"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
}

// GitLabMergeRequestAttributes represents MR attributes
type GitLabMergeRequestAttributes struct {
	ID              int64      `json:"id"`
	IID             int64      `json:"iid"`
	TargetBranch    string     `json:"target_branch"`
	SourceBranch    string     `json:"source_branch"`
	SourceProjectID int64      `json:"source_project_id"`
	AuthorID        int64      `json:"author_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	State           string     `json:"state"` // opened, closed, merged
	MergeStatus     string     `json:"merge_status"`
	URL             string     `json:"url"`
	Action          string     `json:"action"` // open, update, merge, close, reopen
	CreatedAt       GitLabTime  `json:"created_at"`
	UpdatedAt       GitLabTime  `json:"updated_at"`
	MergedAt        *GitLabTime `json:"merged_at"`
	ClosedAt        *GitLabTime `json:"closed_at"`
	LastCommit      GitLabCommit `json:"last_commit"`
}

// GitLabCommit represents a commit
type GitLabCommit struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp GitLabTime `json:"timestamp"`
	URL       string    `json:"url"`
	Author    GitLabCommitAuthor `json:"author"`
}

// GitLabCommitAuthor represents commit author
type GitLabCommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitLabLabel represents a label
type GitLabLabel struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
}

// GitLabMRChanges represents changes in MR
type GitLabMRChanges struct {
	UpdatedAt GitLabValueChange `json:"updated_at"`
}

// GitLabValueChange represents a value change
type GitLabValueChange struct {
	Previous interface{} `json:"previous"`
	Current  interface{} `json:"current"`
}

// GitLabRepository represents repository info
type GitLabRepository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
}

// ShouldTriggerReview determines if this MR event should trigger a review
func (e *GitLabMergeRequestEvent) ShouldTriggerReview() bool {
	// Only trigger review for opened or updated MRs
	action := e.ObjectAttributes.Action
	state := e.ObjectAttributes.State

	// Trigger on: open, update
	// Skip on: merge, close, reopen
	return (action == "open" || action == "update") && state == "opened"
}

// GetMRID returns the merge request IID (project-scoped ID)
func (e *GitLabMergeRequestEvent) GetMRID() int64 {
	return e.ObjectAttributes.IID
}

// GetProjectID returns the project ID
func (e *GitLabMergeRequestEvent) GetProjectID() int64 {
	return e.Project.ID
}

// GetMRTitle returns the MR title
func (e *GitLabMergeRequestEvent) GetMRTitle() string {
	return e.ObjectAttributes.Title
}

// GetMRAuthor returns the author username
func (e *GitLabMergeRequestEvent) GetMRAuthor() string {
	return e.User.Username
}

// GetSourceBranch returns the source branch
func (e *GitLabMergeRequestEvent) GetSourceBranch() string {
	return e.ObjectAttributes.SourceBranch
}

// GetTargetBranch returns the target branch
func (e *GitLabMergeRequestEvent) GetTargetBranch() string {
	return e.ObjectAttributes.TargetBranch
}

// GetMRWebURL returns the MR web URL
func (e *GitLabMergeRequestEvent) GetMRWebURL() string {
	return e.ObjectAttributes.URL
}
