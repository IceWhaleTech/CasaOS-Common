package exec

import (
	"context"
	"testing"
)

func TestCommandContext(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
		args []string
	}
	tests := []struct {
		name string
		args args
		err  bool
	}{
		{
			name: "Test CommandContext",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l"},
			},
			err: false,
		},
		{
			name: "Test CommandContext with noescape",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l", "whoami"},
			},
			err: false,
		},
		{
			name: "Test CommandContext with error",
			args: args{
				ctx:  context.Background(),
				name: "`whoami`",
				args: []string{"-l", "/test"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with injection",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l", "`whoami`"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with multiple injection",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l", "; whoami"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with multiple injection 2",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l", ";", "whoami"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with injection on name",
			args: args{
				ctx:  context.Background(),
				name: "ls ;whoami",
				args: []string{"-l"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with injection on arg-1",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l;whoami"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with quotation injection",
			args: args{
				ctx:  context.Background(),
				name: "ls",
				args: []string{"-l \" \"whoami"},
			},
			err: true,
		},
		{
			name: "Test CommandContext with injection shell script",
			args: args{
				ctx:  context.Background(),
				name: "source /etc/local-storage-helper.sh ;USB_Stop_Auto",
			},
			err: true,
		},
		{
			name: "Test CommandContext with injection shell script divided args",
			args: args{
				ctx:  context.Background(),
				name: "source",
				args: []string{
					"/etc/local-storage-helper.sh",
					";",
					"USB_Stop_Auto",
				},
			},
			err: true,
		},
		{
			name: "Test CommandContext with new-line injection shell script",
			args: args{
				ctx:  context.Background(),
				name: "source /etc/local-storage-helper.sh\nenv",
			},
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommandContext(tt.args.ctx, tt.args.name, tt.args.args...); tt.err != (got.Err != nil) {
				t.Errorf("CommandContext() = %v, want %v", got, tt.err)
			}
		})
	}
}
