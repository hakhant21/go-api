package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"github.com/hakhant21/go-starter/internal/model"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&model.User{},
		&model.RefreshToken{},
		&model.OneTimeToken{},
		&model.Role{},
		&model.Permission{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
