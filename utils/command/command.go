package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
)

// Deprecated: This method is not safe, sould have ensure input.
func OnlyExec(cmdStr string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	fmt.Println(cmd.String())
	buf, err := cmd.CombinedOutput()
	return string(buf), err
}

func ExecResultStr(cmdStr string) (string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)
	fmt.Printf("Executing command: %s\n", cmd.String())

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return "", err
	}

	return buf.String(), cmd.Wait()
}

func ExecResultStrArray(cmdStr string) ([]string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)
	fmt.Printf("Executing command: %s\n", cmd.String())

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var result []string
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, cmd.Wait()
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

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()

		fmt.Printf("Command stderr: %s\n", cmd.Stderr)

		if err != nil {
			fmt.Printf("Failed to execute post-start script %s: %s\n", scriptFilepath, err.Error())
			return err
		}

	}
	fmt.Println("Finished executing post-start scripts.")

	return nil
}

func ExecStdin(stdinStr string, name string, args ...string) error {
	cmd := exec.Command(name, args...)

	fmt.Printf("Executing command: %s\n", cmd.String())

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if _, err := io.WriteString(stdin, stdinStr); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
