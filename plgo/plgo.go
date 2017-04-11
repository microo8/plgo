package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printUsage() {
	fmt.Println(`Usage:
plgo build [path/to/package]
or
plgo install [path/to/package]`)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func buildPackage(buildPath, packageName string) error {
	goBuild := exec.Command("go", "build", "-buildmode=c-shared",
		"-o", filepath.Join(buildPath, packageName+".so"),
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
	if len(flag.Args()) == 0 || len(flag.Args()) > 2 || (flag.Arg(0) != "build" && flag.Arg(0) != "install") {
		printUsage()
		return
	}
	//set package path
	packagePath := "."
	if len(flag.Args()) == 2 {
		packagePath = flag.Arg(1)
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
	err = buildPackage(tempPackagePath, moduleWriter.PackageName)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch flag.Arg(0) {
	case "build":
		err = os.Mkdir("build", 0744)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = copyFile(
			filepath.Join(tempPackagePath, moduleWriter.PackageName+".so"),
			filepath.Join("build", moduleWriter.PackageName+".so"),
		)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = copyFile(
			filepath.Join(tempPackagePath, moduleWriter.PackageName+".sql"),
			filepath.Join("build", moduleWriter.PackageName+".sql"),
		)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "install":
		pglibdirBin, err := exec.Command("pg_config", "--pkglibdir").CombinedOutput()
		if err != nil {
			fmt.Println("Cannot get postgresql libdir:", err)
			return
		}
		pglibdir := strings.TrimSpace(string(pglibdirBin))
		err = copyFile(
			filepath.Join(tempPackagePath, moduleWriter.PackageName+".so"),
			filepath.Join(pglibdir, moduleWriter.PackageName+".so"),
		)
		if err != nil && os.IsPermission(err) {
			cp := exec.Command("sudo", "cp",
				filepath.Join(tempPackagePath, moduleWriter.PackageName+".so"),
				filepath.Join(pglibdir, moduleWriter.PackageName+".so"),
			)
			cp.Stdin = os.Stdin
			cp.Stderr = os.Stderr
			cp.Stdout = os.Stdout
			fmt.Println("Copying the shared objects file to postgres libdir, must get permissions")
			err = cp.Run()
		}
		if err != nil {
			fmt.Println("Cannot copy the shared objects file to postgres libdir:", err)
			return
		}
		//TODO create sql and run it in db
	}
}
