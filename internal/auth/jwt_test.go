package auth

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	id, _ := uuid.NewUUID()
	duration, _ := time.ParseDuration("1m")
	type args struct {
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Basic make JWT test",
			args: args{
				userID:      id,
				tokenSecret: "test",
				expiresIn:   duration,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MakeJWT(tt.args.userID, tt.args.tokenSecret, tt.args.expiresIn)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	id, _ := uuid.NewUUID()
	duration, _ := time.ParseDuration("1m")
	tokenSecret := "test"
	tokenString, _ := MakeJWT(id, tokenSecret, duration)

	type args struct {
		tokenString string
		tokenSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "Basic validate JWT test",
			args: args{
				tokenString: tokenString,
				tokenSecret: tokenSecret,
			},
			want:    id,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateJWT(tt.args.tokenString, tt.args.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateJWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Assert header has bearer token",
			args: args{
				headers: http.Header{"Authorization": []string{"Bearer test"}},
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "Assert header doesn't have bearer token",
			args: args{
				headers: http.Header{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBearerToken(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
