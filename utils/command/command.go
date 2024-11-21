package command

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
)

// Deprecated: This method is not safe, sould have ensure input.
func OnlyExec(cmdStr string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	logger.DebouncedInfo("Executing command: " + cmd.String())
	buf, err := cmd.CombinedOutput()
	return string(buf), err
}

func ExecResultStr(cmdStr string) (string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)

	logger.DebouncedInfo("Executing command: " + cmd.String())

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func ExecResultStrArray(cmdStr string) ([]string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec.Command(cmds[0], cmds[1:]...)

	logger.DebouncedInfo("Executing command: " + cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	result := strings.Split(string(output), "\n")
	return result, nil
}

func ExecuteScripts(scriptDirectory string) error {
	if _, err := os.Stat(scriptDirectory); os.IsNotExist(err) {
		fmt.Printf("No post-start scripts at %s\n", scriptDirectory)
		return err
	}

	files, err := os.ReadDir(scriptDirectory)
	if err != nil {
		fmt.Printf("Failed to read from script directory %s: %s\n", scriptDirectory, err.Error())
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		scriptFilepath := filepath.Join(scriptDirectory, file.Name())

		f, err := os.Open(scriptFilepath)
		if err != nil {
			fmt.Printf("Failed to open script file %s: %s\n", scriptFilepath, err.Error())
			continue
		}
		f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Scan()
		shebang := scanner.Text()

		interpreter := "/bin/sh"
		if strings.HasPrefix(shebang, "#!") {
			interpreter = shebang[2:]
		}

		cmd := exec2.Command(interpreter, scriptFilepath)

		fmt.Printf("Executing post-start script %s using %s\n", scriptFilepath, interpreter)

		_, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to execute post-start script %s: %s\n", scriptFilepath, err.Error())
			return err
		}

	}
	fmt.Println("Finished executing post-start scripts.")

	return nil
}

func ExecStdin(stdinStr string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	logger.DebouncedInfo("Executing command: " + cmd.String())

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	cmd.Stdin = bytes.NewBufferString(stdinStr)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to execute command %s: %s\n", cmd.String(), err.Error())
		return "", err
	}

	return buf.String(), nil
}
