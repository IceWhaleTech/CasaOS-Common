package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
)

// Deprecated: This method is not safe, sould have ensure input.
func OnlyExec(cmdStr string) (string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)
	println(cmd.String())
	buf, err := cmd.CombinedOutput()
	println(string(buf))
	return string(buf), err
}

func ExecResultStr(cmdStr string) (string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)
	println(cmd.String())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return "", err
	}

	buf, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	return string(buf), cmd.Wait()
}

func ExecResultStrArray(cmdStr string) ([]string, error) {
	cmds := strings.Fields(cmdStr)
	cmd := exec2.Command(cmds[0], cmds[1:]...)

	println(cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	buf := []string{}
	outputBuf := bufio.NewReader(stdout)
	for {
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Error :%s\n", err)
			}
			break
		}
		buf = append(buf, string(output))
	}

	return buf, cmd.Wait()
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
		if err != nil {
			fmt.Printf("Failed to execute post-start script %s: %s\n", scriptFilepath, err.Error())
			return err
		}
	}
	fmt.Println("Finished executing post-start scripts.")
	return nil
}
