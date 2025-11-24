package user_test

import (
	"context"
	"errors"
	"testing"

	user "simpleGo/service"
)

func TestCreateGroup(t *testing.T) {
	tests := []struct {
		nameaaa          string
		fu            func() error
		wantErr       bool
		expectedError error
	}{
		{
			nameaaa:    "No Err",
			fu:      func() error { return nil },
			wantErr: false,
		},
		{
			nameaaa:          "Err",
			fu:            func() error { return user.ErrSome },
			wantErr:       true,
			expectedError: user.ErrSome,
		},
	}
	for _, tt := range tests {
		t.Run(tt.nameaaa, func(t *testing.T) {
			gotErr := user.CreateGroup(context.Background(), tt.fu)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateGroup() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				if !errors.Is(gotErr, tt.expectedError) {
					t.Fatalf("unexpected error: got %v, want %v", gotErr, tt.expectedError)
				}
			}
		})
	}
}
