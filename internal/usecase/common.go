package usecase

import (
	"context"
	"embed"
	"fmt"
)

type GetDocsOption struct {
	Lang string
	Name string
}

//go:embed docs/*
var docs embed.FS

func (u Usecase) GetDocs(ctx context.Context, opt GetDocsOption) (string, error) {
	fname := fmt.Sprintf("docs/%s.%s.html", opt.Name, opt.Lang)
	b, err := docs.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
