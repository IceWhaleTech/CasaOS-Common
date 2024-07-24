package command

import "testing"

func TestCommand(t *testing.T) {
	tests := []struct {
		name   string
		cmdStr string
	}{
		
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ExecResultStr(tt.cmdStr)
			if output == "" {
				t.Errorf("ExecResultStr() = %v, want %v", output, tt.cmdStr)
			}

			outputArray := ExecResultStrArray(tt.cmdStr)
			if len(outputArray) == 0 {
				t.Errorf("ExecResultStrArray() = %v, want %v", outputArray, tt.cmdStr)
			}

			OnlyExec(tt.cmdStr)
		})
	}
}
