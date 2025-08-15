package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		comment Comment
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid comment with all fields",
			comment: Comment{
				TaskID:  "task-123",
				Content: "This is a test comment",
			},
			wantErr: false,
		},
		{
			name: "invalid comment - empty task_id",
			comment: Comment{
				Content: "This is a test comment",
			},
			wantErr: true,
			errMsg:  "task_id: task_id is required",
		},
		{
			name: "invalid comment - empty content",
			comment: Comment{
				TaskID: "task-123",
			},
			wantErr: true,
			errMsg:  "content: content is required",
		},
		{
			name: "invalid comment - both empty",
			comment: Comment{},
			wantErr: true,
			errMsg:  "task_id: task_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			
			if tt.wantErr {
				require.Error(t, err, "Expected validation error but got none")
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err, "Unexpected validation error")
			}
		})
	}
}