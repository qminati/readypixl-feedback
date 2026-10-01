package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/bus"
	"github.com/getfider/fider/app/pkg/web"
)

// LlmsTXT serves the whole board as plain markdown so an AI agent can read it
// from one URL (llmstxt.org convention). /llms.txt lists every public post;
// /llms-full.txt (full=true) adds each post's text, staff reply and comments.
func LlmsTXT(full bool) web.HandlerFunc {
	return func(c *web.Context) error {
		if c.Tenant().IsPrivate {
			return c.NotFound()
		}

		posts := &query.SearchPosts{View: "all", Limit: "all"}
		if err := bus.Dispatch(c, posts); err != nil {
			return c.Failure(err)
		}

		baseURL := c.BaseURL()
		helpCenter := os.Getenv("READYPIXL_HELPCENTER_URL")
		if helpCenter == "" {
			helpCenter = "https://readypixl-helpcenter.vercel.app"
		}

		text := strings.Builder{}
		fmt.Fprintf(&text, "# %s\n\n", c.Tenant().Name)
		text.WriteString("> Public ReadyPixl feedback board: feature requests, bug reports and questions, with votes, status and staff replies.")
		if full {
			fmt.Fprintf(&text, " Index: %s/llms.txt\n\n", baseURL)
		} else {
			fmt.Fprintf(&text, " Full text of every post with replies and comments: %s/llms-full.txt\n\n", baseURL)
		}
		fmt.Fprintf(&text, "- Help center (how ReadyPixl works): %s/llms.txt\n", helpCenter)
		fmt.Fprintf(&text, "- JSON API, no sign-in needed to read: %s/api/v1/posts?view=all&limit=all, %s/api/v1/posts/{number}, %s/api/v1/posts/{number}/comments, %s/api/v1/similarposts?query=...\n", baseURL, baseURL, baseURL, baseURL)
		text.WriteString("- Statuses: open, planned, started, completed, declined, duplicate\n\n")

		count := 0
		for _, post := range posts.Result {
			if !post.IsApproved {
				continue
			}
			count++
			if full {
				writeFullPost(c, &text, post, baseURL)
			} else {
				fmt.Fprintf(&text, "- [#%d %s](%s): %s, %d votes, %d comments\n",
					post.Number, post.Title, post.Url(baseURL), post.Status.Name(), post.VotesCount, post.CommentsCount)
			}
		}
		if count == 0 {
			text.WriteString("No posts yet.\n")
		}

		return c.String(http.StatusOK, text.String())
	}
}

func writeFullPost(c *web.Context, text *strings.Builder, post *entity.Post, baseURL string) {
	fmt.Fprintf(text, "---\n\n## #%d %s\n\n", post.Number, post.Title)
	fmt.Fprintf(text, "URL: %s\n", post.Url(baseURL))
	fmt.Fprintf(text, "Status: %s | Votes: %d | Comments: %d | Posted: %s", post.Status.Name(), post.VotesCount, post.CommentsCount, post.CreatedAt.Format("2006-01-02"))
	if post.User != nil {
		fmt.Fprintf(text, " by %s", post.User.Name)
	}
	text.WriteString("\n")
	if len(post.Tags) > 0 {
		fmt.Fprintf(text, "Tags: %s\n", strings.Join(post.Tags, ", "))
	}
	if strings.TrimSpace(post.Description) != "" {
		fmt.Fprintf(text, "\n%s\n", strings.TrimSpace(post.Description))
	}
	if post.Response != nil && strings.TrimSpace(post.Response.Text) != "" {
		fmt.Fprintf(text, "\n### ReadyPixl reply (%s)\n\n%s\n", post.Response.RespondedAt.Format("2006-01-02"), strings.TrimSpace(post.Response.Text))
	}
	if post.CommentsCount > 0 {
		comments := &query.GetCommentsByPost{Post: post}
		if err := bus.Dispatch(c, comments); err == nil {
			text.WriteString("\n### Comments\n\n")
			for _, comment := range comments.Result {
				if !comment.IsApproved {
					continue
				}
				author := "someone"
				if comment.User != nil {
					author = comment.User.Name
				}
				content := entity.CommentString(comment.Content).SanitizeMentions()
				fmt.Fprintf(text, "- %s (%s): %s\n", author, comment.CreatedAt.Format("2006-01-02"), strings.ReplaceAll(strings.TrimSpace(content), "\n", "\n  "))
			}
		}
	}
	text.WriteString("\n")
}
