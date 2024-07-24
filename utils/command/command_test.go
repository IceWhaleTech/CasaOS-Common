package command

import (
	"os"
	"testing"
)

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
		// TODO: Fix this test
		// {
		// 	name:   "Test Command with noescape",
		// 	cmdStr: "ls -l whoami",
		// 	noErr:  true,
		// },
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
			cmdStr: "ls ;whoami -l",
		},
		{
			name:   "Test Command with injection on arg-1",
			cmdStr: "ls -l;whoami",
		},
		{
			name:   "Test Command with quotation injection",
			cmdStr: "ls -l \" \"whoami",
		},
		{
			name:   "Test Command with injection shell script",
			cmdStr: "source /etc/local-storage-helper.sh ;USB_Stop_Auto",
		},
		{
			name:   "Test Command with injection shell script divided args",
			cmdStr: "source /etc/local-storage-helper.sh ; USB_Stop_Auto",
		},
		{
			name:   "Test Command with new-line injection shell script",
			cmdStr: "source /etc/local-storage-helper.sh\nenv",
		},
	}

	t.Run("TestExecuteScripts", func(t *testing.T) {
		// make a temp directory
		tmpDir, err := os.MkdirTemp("", "casaos-test-*")
		if err != nil {
			t.Error(err)
		}
		defer os.RemoveAll(tmpDir)

		ExecuteScripts(tmpDir)

		// create a sample script under tmpDir
		script := tmpDir + "/test.sh"
		f, err := os.Create(script)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()

		// write a sample script
		_, err = f.WriteString("#!/bin/bash\necho 123")
		if err != nil {
			t.Error(err)
		}

		ExecuteScripts(tmpDir)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if output, err := OnlyExec(tt.cmdStr); tt.noErr != (err == nil) {
				t.Errorf("OnlyExec() error = %v, wantErr %v", err, tt.noErr)
			} else {
				t.Logf("Output: %s", output)
			}
		})

		t.Run(tt.name, func(t *testing.T) {
			if output, err := ExecResultStr(tt.cmdStr); tt.noErr != (err == nil) {
				t.Errorf("ExecResultStr() error = %v, wantErr %v", err, tt.noErr)
			} else {
				t.Logf("Output: %s", output)
			}
		})

		t.Run(tt.name, func(t *testing.T) {
			if output, error := ExecResultStrArray(tt.cmdStr); tt.noErr != (error == nil) {
				t.Errorf("ExecResultStrArray() error = %v, wantErr %v", error, tt.noErr)
			} else {
				t.Logf("Output: %v", output)
			}
		})
	}
}
