package exec

import (
	"testing"
)

func TestCommand(t *testing.T) {
	type args struct {
		name string
		args []string
	}
	tests := []struct {
		name  string
		args  args
		noErr bool
	}{
		{
			name: "Test Command",
			args: args{
				name: "ls",
				args: []string{"-l"},
			},
			noErr: true,
		},
		{
			name: "Test Command with noescape",
			args: args{
				name: "ls",
				args: []string{"-l", "whoami"},
			},
			noErr: true,
		},
		{
			name: "Test Command with error",
			args: args{
				name: "`whoami`",
				args: []string{"-l", "/test"},
			},
		},
		{
			name: "Test Command with injection",
			args: args{
				name: "ls",
				args: []string{"-l", "`whoami`"},
			},
		},
		{
			name: "Test Command with multiple injection",
			args: args{
				name: "ls",
				args: []string{"-l", "; whoami"},
			},
		},
		{
			name: "Test Command with multiple injection 2",
			args: args{
				name: "ls",
				args: []string{"-l", ";", "whoami"},
			},
		},
		{
			name: "Test Command with injection on name",
			args: args{
				name: "ls ;whoami",
				args: []string{"-l"},
			},
		},
		{
			name: "Test Command with injection on arg-1",
			args: args{
				name: "ls",
				args: []string{"-l;whoami"},
			},
		},
		{
			name: "Test Command with quotation injection",
			args: args{
				name: "ls",
				args: []string{"-l \" \"whoami"},
			},
		},
		{
			name: "Test Command with injection shell script",
			args: args{
				name: "source /etc/local-storage-helper.sh ;USB_Stop_Auto",
			},
		},
		{
			name: "Test Command with injection shell script divided args",
			args: args{
				name: "source",
				args: []string{"/etc/local-storage-helper.sh", ";", "USB_Stop_Auto"},
			},
		},
		{
			name: "Test Command with new-line injection shell script",
			args: args{
				name: "source /etc/local-storage-helper.sh\nenv",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Command(tt.args.name, tt.args.args...); tt.noErr != (got.Err == nil) {
				t.Errorf("Command() = %v, want %v", got, tt.noErr)
			}
		})
	}
}
