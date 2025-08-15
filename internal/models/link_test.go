package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLink_Validate(t *testing.T) {
	tests := []struct {
		name    string
		link    Link
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid link with all fields",
			link: Link{
				TaskID: "task-123",
				Type:   PullRequest,
				URL:    "https://github.com/org/repo/pull/123",
				Title:  "Fix memory leak",
				Status: "open",
			},
			wantErr: false,
		},
		{
			name: "valid link with minimal fields - title defaulted to URL",
			link: Link{
				TaskID: "task-123",
				Type:   SlackThread,
				URL:    "https://slack.com/thread/123",
			},
			wantErr: false,
		},
		{
			name: "invalid link - empty task_id",
			link: Link{
				Type:  PullRequest,
				URL:   "https://github.com/org/repo/pull/123",
				Title: "Fix memory leak",
			},
			wantErr: true,
			errMsg:  "task_id: task_id is required",
		},
		{
			name: "invalid link - empty URL",
			link: Link{
				TaskID: "task-123",
				Type:   PullRequest,
				Title:  "Fix memory leak",
			},
			wantErr: true,
			errMsg:  "url: url is required",
		},
		{
			name: "invalid link - invalid type",
			link: Link{
				TaskID: "task-123",
				Type:   LinkType("invalid"),
				URL:    "https://github.com/org/repo/pull/123",
			},
			wantErr: true,
			errMsg:  "type: invalid link type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.link.Validate()
			
			if tt.wantErr {
				require.Error(t, err, "Expected validation error but got none")
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err, "Unexpected validation error")
				
				// Check title defaults to URL if empty
				if tt.name == "valid link with minimal fields - title defaulted to URL" {
					assert.Equal(t, tt.link.URL, tt.link.Title)
				}
			}
		})
	}
}

func TestLinkType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		linkType LinkType
		want     bool
	}{
		{"valid pull_request", PullRequest, true},
		{"valid slack_thread", SlackThread, true},
		{"valid jira_ticket", JiraTicket, true},
		{"valid documentation", Documentation, true},
		{"valid other", Other, true},
		{"invalid empty", LinkType(""), false},
		{"invalid random", LinkType("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.linkType.IsValid())
		})
	}
}

func TestLinkConstants(t *testing.T) {
	// Test that link type constants have expected values
	assert.Equal(t, LinkType("pull_request"), PullRequest)
	assert.Equal(t, LinkType("slack_thread"), SlackThread)
	assert.Equal(t, LinkType("jira_ticket"), JiraTicket)
	assert.Equal(t, LinkType("documentation"), Documentation)
	assert.Equal(t, LinkType("other"), Other)
	
	// Test that all constants are valid
	validTypes := []LinkType{PullRequest, SlackThread, JiraTicket, Documentation, Other}
	for _, linkType := range validTypes {
		assert.True(t, linkType.IsValid(), "LinkType %s should be valid", linkType)
	}
}