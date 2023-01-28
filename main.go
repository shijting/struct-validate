package main

import (
	"SJT/struct-validate/pkg"
	"SJT/struct-validate/utils"
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
)

func main() {
	cmd := &cobra.Command{}
	cmd.AddCommand(GenerateCmd())
	cmd.Execute()
}

func GenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "validate",
		Short:   "generate validate code for the directory",
		Example: "struct-validate validate .",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				panic("参数不正确")
			}
			dir, _ := filepath.Abs(args[0])
			fmt.Println(dir, ":")
			files, err := utils.ScanFiles(args[0])
			if err != nil {
				panic(err)
			}
			s := pkg.ScanFile{Files: files}
			err = s.Resolver()
			if err != nil {
				//panic(err)
				fmt.Println(err)
			}
		},
	}
}
