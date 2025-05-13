package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordChangeRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request PasswordChangeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
			request: PasswordChangeRequest{
				CurrentPassword:    "oldpassword",
				NewPassword:        "newpassword123",
				ConfirmNewPassword: "newpassword123",
			},
			wantErr: false,
		},
		{
			name: "Missing current password",
			request: PasswordChangeRequest{
				CurrentPassword:    "",
				NewPassword:        "newpassword123",
				ConfirmNewPassword: "newpassword123",
			},
			wantErr: true,
			errMsg:  "field is required",
		},
		{
			name: "Missing new password",
			request: PasswordChangeRequest{
				CurrentPassword:    "oldpassword",
				NewPassword:        "",
				ConfirmNewPassword: "newpassword123",
			},
			wantErr: true,
			errMsg:  "field is required",
		},
		{
			name: "Missing confirm password",
			request: PasswordChangeRequest{
				CurrentPassword:    "oldpassword",
				NewPassword:        "newpassword123",
				ConfirmNewPassword: "",
			},
			wantErr: true,
			errMsg:  "field is required",
		},
		{
			name: "Passwords don't match",
			request: PasswordChangeRequest{
				CurrentPassword:    "oldpassword",
				NewPassword:        "newpassword123",
				ConfirmNewPassword: "different",
			},
			wantErr: true,
			errMsg:  "passwords do not match",
		},
		{
			name: "Password too short",
			request: PasswordChangeRequest{
				CurrentPassword:    "oldpassword",
				NewPassword:        "short",
				ConfirmNewPassword: "short",
			},
			wantErr: true,
			errMsg:  "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
