package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var newCMD = &cobra.Command{
	Use:   "new",
	Short: "Generate mainstay of golang server project",
	Run: func(cmd *cobra.Command, args []string) {
		newProj(args)
	},
}

var circle bool

func init() {
	newCMD.PersistentFlags().BoolVarP(&circle, "circle", "c", false, "generate circle.yml")
	rootCmd.AddCommand(newCMD)
}

// newProj generate a new project
func newProj(args []string) {
	if len(args) == 0 {
		log.Fatal("Missing project name")
	}
	name := args[0]
	log.Printf("Generate project: %s", name)

	if err := os.Mkdir(name, 0755); err != nil {
		log.Fatal(err)
	}

	if err := genREADME(name); err != nil {
		log.Fatal(err)
	}

	if err := genMain(name); err != nil {
		log.Fatal(err)
	}

	if err := genConf(name); err != nil {
		log.Fatal(err)
	}

	if circle {
		if err := genCircleYml(name); err != nil {
			log.Fatal(err)
		}
	}

	if err := genIgnore(name); err != nil {
		log.Fatal(err)
	}

	if err := genGoGenerate(name); err != nil {
		log.Fatal(err)
	}

	if err := genMakefile(name); err != nil {
		log.Fatal(err)
	}

	if err := genProto(name); err != nil {
		log.Fatal(err)
	}
}

func genREADME(name string) error {
	path := filepath.Join(name, "README.md")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintln(f, `#`, name); err != nil {
		return err
	}
	return nil
}

func genMain(name string) error {

	path := filepath.Join(name, "main.go")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "package main")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, `import (
	"log"
	
	"github.com/philchia/zen"
)`)
	fmt.Fprintln(f, "")

	fmt.Fprintln(f, "func main() {")
	fmt.Fprintln(f, "	server := zen.New()")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, `	if err := server.Run(":8080"); err != nil {`)
	fmt.Fprintln(f, "		log.Fatal(err)")
	fmt.Fprintln(f, "	}")
	fmt.Fprintln(f, "}")

	return nil
}

func genConf(name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	path := filepath.Join(name, "conf")
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}
	os.Chdir(path)
	f, err := os.OpenFile("conf.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "package conf")

	return nil
}

func genGoGenerate(name string) error {
	path := filepath.Join(name, "zgen.go")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "package main")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, `//go:generate make -s -f generate/makefile`)

	return nil
}

func genMakefile(name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	path := filepath.Join(name, "generate")
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}
	os.Chdir(path)
	f, err := os.OpenFile("makefile", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "gen_proto: proto/*.pb.go")
	fmt.Fprintln(f, `
clean_proto:
	rm -f proto/*.pb.go`)
	fmt.Fprintln(f, `
proto/*.pb.go: proto/*.proto $(shell which protoc) $(shell which protoc-gen-go)
	@echo "gen proto"
	@rm -f proto/*.pb.go
	@protoc --go_out=plugins=grpc:. proto/*.proto`)

	return nil
}

func genTests(name string) error {
	return nil
}

func genIgnore(name string) error {
	path := filepath.Join(name, ".gitignore")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, name)
	fmt.Fprintln(f, ignoreTmpl)

	return nil
}

func genCircleYml(name string) error {
	path := filepath.Join(name, "circle.yml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, name)
	fmt.Fprintln(f, circleTmpl)

	return nil
}

func genProto(name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	path := filepath.Join(name, "proto")
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}
	os.Chdir(path)
	f, err := os.OpenFile(fmt.Sprintf("%s.proto", name), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, `syntax = "proto3";
package proto;

service %s {
	rpc Ping(Ping) returns (PingResp);
}

message Ping {
	string Msg = 1;
}

message PingResp {
	string Msg = 1;
}`, name)

	return nil
}
