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

func OnlyExec(cmdStr string) error {
	cmd := exec2.Command("/bin/bash", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return err
	}

	return cmd.Wait()
}

func ExecResultStrArray(cmdStr string) ([]string, error) {
	cmd := exec2.Command("/bin/bash", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	networklist := []string{}
	outputBuf := bufio.NewReader(stdout)
	for {
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Error :%s\n", err)
			}
			break
		}
		networklist = append(networklist, string(output))
	}

	return networklist, cmd.Wait()
}

func ExecResultStr(cmdStr string) (string, error) {
	cmd := exec2.Command("/bin/bash", "-c", cmdStr)
	println(cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		return "", err
	}
	str, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(str), cmd.Wait()
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
