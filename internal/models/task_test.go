package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task with all fields",
			task: Task{
				Title:    "Test task",
				JiraID:   "OCPBUGS-1234",
				Priority: High,
				Status:   InProgress,
			},
			wantErr: false,
		},
		{
			name: "valid task with minimal fields - defaults applied",
			task: Task{
				Title: "Test task",
			},
			wantErr: false,
		},
		{
			name: "invalid task - empty title",
			task: Task{
				JiraID:   "OCPBUGS-1234",
				Priority: High,
				Status:   InProgress,
			},
			wantErr: true,
			errMsg:  "title: title is required",
		},
		{
			name: "invalid priority",
			task: Task{
				Title:    "Test task",
				Priority: Priority("invalid"),
				Status:   InProgress,
			},
			wantErr: true,
			errMsg:  "priority: invalid priority",
		},
		{
			name: "invalid status",
			task: Task{
				Title:  "Test task",
				Status: Status("invalid"),
			},
			wantErr: true,
			errMsg:  "status: invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				
				// Check defaults are applied for minimal task
				if tt.name == "valid task with minimal fields - defaults applied" {
					assert.Equal(t, DefaultNoJira, tt.task.JiraID)
					assert.Equal(t, DefaultPriority, tt.task.Priority)
					assert.Equal(t, DefaultStatus, tt.task.Status)
				}
			}
		})
	}
}

func TestTask_ValidateDefaults(t *testing.T) {
	task := Task{
		Title: "Test task",
	}
	
	err := task.Validate()
	require.NoError(t, err)
	
	assert.Equal(t, DefaultNoJira, task.JiraID)
	assert.Equal(t, DefaultPriority, task.Priority)
	assert.Equal(t, DefaultStatus, task.Status)
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     bool
	}{
		{"valid minor", Minor, true},
		{"valid normal", Normal, true},
		{"valid high", High, true},
		{"valid critical", Critical, true},
		{"invalid empty", Priority(""), false},
		{"invalid random", Priority("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.priority.IsValid())
		})
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"valid new", New, true},
		{"valid in_progress", InProgress, true},
		{"valid blocked", Blocked, true},
		{"valid done", Done, true},
		{"valid archived", Archived, true},
		{"invalid empty", Status(""), false},
		{"invalid random", Status("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}
	
	expected := "test_field: test message"
	assert.Equal(t, expected, err.Error())
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	assert.Equal(t, Priority("normal"), DefaultPriority)
	assert.Equal(t, Status("new"), DefaultStatus)
	assert.Equal(t, "NO-JIRA", DefaultNoJira)
	
	// Test that defaults are valid
	assert.True(t, DefaultPriority.IsValid())
	assert.True(t, DefaultStatus.IsValid())
}