package categorizer

import (
	"regexp"
	"strings"
	"stak/internal/models"
)

type Categoriser struct {
	linkRegex     *regexp.Regexp
	todoRegex     *regexp.Regexp
	codeRegex     *regexp.Regexp
	questionRegex *regexp.Regexp
	meetingRegex  *regexp.Regexp
}

func New() *Categoriser {
	return &Categoriser{
		linkRegex:     regexp.MustCompile(`https?://[^\s]+`),
		todoRegex:     regexp.MustCompile(`(?i)^(\s*-\s*\[\s*\]\s*|todo:|\[\s*\]|\*\s+|•\s+)`),
		codeRegex:     regexp.MustCompile("```|`[^`]+`|\\$\\s+[a-zA-Z]|import\\s+|function\\s+|class\\s+|def\\s+|const\\s+|let\\s+|var\\s+"),
		questionRegex: regexp.MustCompile(`\?(\s|$)`),
		meetingRegex:  regexp.MustCompile(`(?i)(meeting|standup|sync|1:1|one-on-one|call|zoom)`),
	}
}

func (c *Categoriser) CategoriseEntry(entry *models.Entry) {
	content := strings.ToLower(entry.Content)
	originalContent := entry.Content

	switch {
	case c.linkRegex.MatchString(originalContent):
		entry.Type = models.TypeLink
		if urls := c.linkRegex.FindAllString(originalContent, -1); len(urls) > 0 {
			entry.URL = urls[0]
		}
		c.extractTags(entry, []string{"link", "web", "reference"})

	case c.todoRegex.MatchString(originalContent):
		entry.Type = models.TypeTodo
		entry.TodoStatus = models.TodoPending
		c.extractTags(entry, []string{"todo", "task"})

	case c.codeRegex.MatchString(originalContent):
		entry.Type = models.TypeCode
		c.extractCodeTags(entry, originalContent)

	case c.questionRegex.MatchString(originalContent):
		entry.Type = models.TypeQuestion
		c.extractTags(entry, []string{"question", "inquiry"})
		c.extractLanguageTags(entry, originalContent)

	case c.meetingRegex.MatchString(content):
		entry.Type = models.TypeMeeting
		c.extractTags(entry, []string{"meeting", "discussion"})

	default:
		entry.Type = models.TypeNote
		c.extractGeneralTags(entry, content)
	}

	c.extractDomainTags(entry, content)
}

func (c *Categoriser) extractCodeTags(entry *models.Entry, content string) {
	tags := []string{"code"}
	
	languages := map[string]string{
		"go":         "golang",
		"golang":     "golang",
		"javascript": "js",
		"typescript": "ts",
		"python":     "python",
		"rust":       "rust",
		"java":       "java",
		"docker":     "docker",
		"sql":        "database",
		"bash":       "shell",
		"yaml":       "config",
		"json":       "config",
	}

	contentLower := strings.ToLower(content)
	for keyword, tag := range languages {
		if strings.Contains(contentLower, keyword) {
			tags = append(tags, tag)
		}
	}
	
	c.extractTags(entry, tags)
}

func (c *Categoriser) extractGeneralTags(entry *models.Entry, content string) {
	keywords := map[string]string{
		"idea":    "idea",
		"brainstorm": "brainstorm",
		"thought": "reflection",
		"note":    "note",
		"reminder": "reminder",
		"important": "important",
		"urgent":   "urgent",
		"bug":      "bug",
		"feature":  "feature",
		"fix":      "fix",
	}

	var tags []string
	for keyword, tag := range keywords {
		if strings.Contains(content, keyword) {
			tags = append(tags, tag)
		}
	}
	
	// Also check for programming languages in all content
	c.extractLanguageTags(entry, content)
	
	if len(tags) == 0 {
		tags = []string{"note"}
	}
	
	c.extractTags(entry, tags)
}

func (c *Categoriser) extractLanguageTags(entry *models.Entry, content string) {
	languages := map[string]string{
		"go":         "golang",
		"golang":     "golang",
		"javascript": "js",
		"typescript": "ts",
		"python":     "python",
		"rust":       "rust",
		"java":       "java",
		"docker":     "docker",
		"sql":        "database",
		"bash":       "shell",
		"yaml":       "config",
		"json":       "config",
	}

	contentLower := strings.ToLower(content)
	for keyword, tag := range languages {
		if strings.Contains(contentLower, keyword) {
			c.extractTags(entry, []string{tag})
		}
	}
}

func (c *Categoriser) extractDomainTags(entry *models.Entry, content string) {
	domains := map[string]string{
		"work":     "work",
		"personal": "personal",
		"project":  "project",
		"learning": "learning",
		"research": "research",
		"client":   "client",
		"team":     "team",
	}

	for keyword, tag := range domains {
		if strings.Contains(content, keyword) {
			entry.Tags = append(entry.Tags, tag)
		}
	}
}

func (c *Categoriser) extractTags(entry *models.Entry, tags []string) {
	for _, tag := range tags {
		if !contains(entry.Tags, tag) {
			entry.Tags = append(entry.Tags, tag)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}