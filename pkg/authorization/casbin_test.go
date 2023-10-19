package authorization

import "testing"

func TestFormatResourceArn(t *testing.T) {
	type args struct {
		arn string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test format resource arn",
			args: args{
				arn: "aws:s3:::my_corporate_bucket/exampleobject.png",
			},
			want: "aws:s3:::my_corporate_bucket/exampleobject.png",
		},
		{
			name: "test format resource arn",
			args: args{
				arn: "app:tenant:model:field1/*:field2/value2",
			},
			want: "app:tenant:model:field2/value2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatResourceArn(tt.args.arn); got != tt.want {
				t.Errorf("FormatResourceArn() = %v, want %v", got, tt.want)
			}
		})
	}
}
