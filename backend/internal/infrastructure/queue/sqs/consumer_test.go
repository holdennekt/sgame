package sqs

import (
	"reflect"
	"testing"
)

func TestParseS3EventKeys(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "object created with encoded key",
			raw:  `{"Records":[{"eventName":"ObjectCreated:Post","s3":{"object":{"key":"drafts/a+b%2Bc.png"}}}]}`,
			want: []string{"drafts/a b+c.png"},
		},
		{
			name: "non-create events are skipped",
			raw:  `{"Records":[{"eventName":"ObjectRemoved:Delete","s3":{"object":{"key":"x"}}}]}`,
		},
		{
			name: "test event has no keys",
			raw:  `{"Service":"Amazon S3","Event":"s3:TestEvent","Bucket":"b"}`,
		},
		{
			name:    "garbage is an error",
			raw:     `not json`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseS3EventKeys([]byte(tt.raw))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keys = %v, want %v", got, tt.want)
			}
		})
	}
}
