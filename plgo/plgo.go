package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func printUsage() {
	fmt.Println(`Usage: plgo [-v] [path/to/package]`)
	flag.PrintDefaults()
}

func buildPackage(buildPath, packageName string) error {
	if err := os.Setenv("CGO_LDFLAGS_ALLOW", "-shared"); err != nil {
		return err
	}
	switchx := "-v" // substitutor
	if verbose {
		switchx = "-x"
	}
	fileExt := ".so"
	if runtime.GOOS == "windows" {
		fileExt = ".dll"
	}
	goBuild := exec.Command("go", "build", switchx,
		"-buildmode=c-shared",
		"-o", filepath.Join("build", packageName+fileExt),
		filepath.Join(buildPath, "package.go"),
		filepath.Join(buildPath, "methods.go"),
		filepath.Join(buildPath, "pl.go"),
	)
	goBuild.Stdout = os.Stdout
	goBuild.Stderr = os.Stderr
	if err := goBuild.Run(); err != nil {
		return fmt.Errorf("Cannot build package: %s", err)
	}
	return nil
}

var verbose bool

func main() {
	flag.BoolVar(&verbose, "v", false, "be verbose, 'go build -x'")
	flag.Parse()
	packagePath := "."
	if len(flag.Args()) == 1 {
		packagePath = flag.Arg(0)
	}
	moduleWriter, err := NewModuleWriter(packagePath)
	if err != nil {
		fmt.Println(err)
		printUsage()
		return
	}
	tempPackagePath, err := moduleWriter.WriteModule()
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println(tempPackagePath)
	if _, err = os.Stat("build"); os.IsNotExist(err) {
		err = os.Mkdir("build", 0744)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = buildPackage(tempPackagePath, moduleWriter.PackageName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = moduleWriter.WriteSQL("build")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = moduleWriter.WriteControl("build")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = moduleWriter.WriteMakefile("build")
	if err != nil {
		fmt.Println(err)
		return
	}
}
