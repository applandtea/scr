package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

var supportedLanguage = []string{
	"bash",
	"c",
	"cpp",
	"go",
	"haskell",
	"java",
	"perl",
	"php",
	"python",
	"scala",
	"ruby",
	"rust",
}

var cLanguage = []string{
	"c",
	"cpp",
	"go",
	"java",
	"scala",
	"rust",
}

func goRun(workDir, stdin string, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	// args[0] is the program name
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

// Run will run payload, if the language need compile
// it will call CompileAndRun()
func (ar *PayLoad) Run() {

	if len(ar.A.Run) == 0 {
		if ar.L == "haskell" {
			ar.A.Run = []string{"runhaskell"}
		} else {
			ar.A.Run = []string{ar.L}
		}
	}
	args := ar.A.Run[0:]
	var workDir string
	if len(ar.F) != 0 {
		absFilePaths, err := writeFiles(ar.F)
		if err != nil {
			exitF("Write files failed")
		}
		workDir = filepath.Dir(absFilePaths[0])
		args = append(ar.A.Run[0:], absFilePaths...)
	}

	stdOut, stdErr, exitErr := goRun(workDir, ar.I, args...)
	returnStdOut(stdOut, stdErr, errToStr(exitErr))
}

func (ar *PayLoad) compileAndRun() {
	// if no file return
	if len(ar.F) == 0 {
		exitF("No files given")
	}
	absFilePaths, err := writeFiles(ar.F)
	if err != nil {
		exitF("Write files failed")
	}
	workDir := filepath.Dir(absFilePaths[0])

	switch {
	case ar.L == "c" || ar.L == "cpp" || ar.L == "rust":
		if len(ar.A.Compile) == 0 {
			switch ar.L {
			case "c":
				ar.A.Compile = []string{"gcc"}
			case "cpp":
				ar.A.Compile = []string{"g++"}
			case "rust":
				ar.A.Compile = []string{"rustc"}
			}
		}
		binName := "a.out"

		args := append(ar.A.Compile, []string{"-o", binName}...)
		args = append(args, absFilePaths...)
		// compile
		stdOut, stdErr, exitErr := goRun(workDir, "", args...)
		if exitErr != nil {
			if _, ok := exitErr.(*exec.ExitError); ok {
				returnStdOut(stdOut, stdErr, errToStr(exitErr))
				exitF("Compile Error")
			}
			exitF("Ric goRun Failed")
		}

		// run
		binPath := filepath.Join(workDir, binName)
		args = append(ar.A.Run, binPath)

		stdOut, stdErr, exitErr = goRun(workDir, ar.I, args...)
		returnStdOut(stdOut, stdErr, errToStr(exitErr))

	case ar.L == "java" || ar.L == "scala":
		if len(ar.A.Compile) == 0 {
			if ar.L == "java" {
				ar.A.Compile = []string{"javac"}
			} else if ar.L == "scala" {
				ar.A.Compile = []string{"scalac"}
			}
		}

		args := append(ar.A.Compile, absFilePaths...)

		filename := filepath.Base(absFilePaths[0])

		// compile
		stdOut, stdErr, exitErr := goRun(workDir, "", args...)
		if exitErr != nil {
			returnStdOut(stdOut, stdErr, errToStr(exitErr))
			exitF("Compile Error")
		}

		if len(ar.A.Run) == 0 {
			if ar.L == "java" {
				ar.A.Run = []string{"java"}
			} else if ar.L == "scala" {
				ar.A.Run = []string{"scala"}
			}
		}
		args = append(ar.A.Run, javaClassName(filename))
		stdOut, stdErr, exitErr = goRun(workDir, ar.I, args...)
		returnStdOut(stdOut, stdErr, errToStr(exitErr))

	case ar.L == "go":
		if len(ar.A.Compile) == 0 {
			ar.A.Compile = []string{"go", "build"}
		}
		binName := "main"

		args := append(ar.A.Compile, []string{"-o", binName}...)
		args = append(args, absFilePaths...)
		// compile
		stdOut, stdErr, exitErr := goRun(workDir, "", args...)
		if exitErr != nil {
			if _, ok := exitErr.(*exec.ExitError); ok {
				returnStdOut(stdOut, stdErr, errToStr(exitErr))
				exitF("Compile Error")
			}
			exitF("Ric goRun Failed")
		}

		// run
		binPath := filepath.Join(workDir, binName)
		args = append(ar.A.Run, binPath)

		stdOut, stdErr, exitErr = goRun(workDir, ar.I, args...)
		returnStdOut(stdOut, stdErr, errToStr(exitErr))

	default:
		exitF("Unsupported compile language: %s", ar.L)
	}
}

func (ar *PayLoad) needCompile() bool {
	for _, l := range cLanguage {
		if ar.L == l {
			return true
		}
	}
	return false
}

func (ar *PayLoad) isSupport() bool {
	for _, l := range supportedLanguage {
		if ar.L == l {
			return true
		}
	}
	return false
}

func javaClassName(filename string) string {
	ext := filepath.Ext(filename)
	return filename[0 : len(filename)-len(ext)]
}
