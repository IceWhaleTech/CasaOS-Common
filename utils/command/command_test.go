package command

import "testing"

func TestCommand(t *testing.T) {
	tests := []struct {
		name   string
		cmdStr string
		noErr  bool
	}{
		{
			name:   "Test Command",
			cmdStr: "ls -l",
			noErr:  true,
		},
		{
			name:   "Test Command with noescape",
			cmdStr: "ls -l whoami",
			noErr:  true,
		},
		{
			name:   "Test Command with error",
			cmdStr: "`whoami` -l /test",
		},
		{
			name:   "Test Command with injection",
			cmdStr: "ls -l `whoami`",
		},
		{
			name:   "Test Command with multiple injection",
			cmdStr: "ls -l ; whoami",
		},
		{
			name:   "Test Command with multiple injection 2",
			cmdStr: "ls -l ; whoami",
		},
		{
			name:   "Test Command with injection on name",
			cmdStr: "ls -l;whoami",
		},
		{
			name:   "Test Command with injection on arg-1",
			cmdStr: "ls -l;whoami",
		},
		{
			name:   "Test Command with quotation injection",
			cmdStr: "ls -l;\"whoami\"",
		},
		{
			name:   "Test Command with injection shell script",
			cmdStr: "source /etc/local-storage-helper.sh;USB_Stop_Auto",
		},
		{
			name:   "Test Command with injection shell script divided args",
			cmdStr: "source /etc/local-storage-helper.sh ; USB_Stop_Auto",
		},
		{
			name:   "Test Command with new-line injection shell script",
			cmdStr: "source /etc/local-storage-helper.sh\nUSB_Stop_Auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := OnlyExec(tt.cmdStr); tt.noErr != (err == nil) {
				t.Errorf("OnlyExec() error = %v, wantErr %v", err, tt.noErr)
			}
		})

		t.Run(tt.name, func(t *testing.T) {
		})

		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
