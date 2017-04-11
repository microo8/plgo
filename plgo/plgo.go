package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func printUsage() {
	fmt.Println(`Usage: plgo [path/to/package]`)
}

func buildPackage(buildPath, packageName string) error {
	goBuild := exec.Command("go", "build", "-buildmode=c-shared",
		"-o", filepath.Join("build", packageName+".so"),
		filepath.Join(buildPath, "package.go"),
		filepath.Join(buildPath, "methods.go"),
		filepath.Join(buildPath, "pl.go"),
	)
	goBuild.Stdout = os.Stdout
	goBuild.Stderr = os.Stderr
	err := goBuild.Run()
	if err != nil {
		return fmt.Errorf("Cannot build package: %s", err)
	}
	return nil
}

func main() {
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
