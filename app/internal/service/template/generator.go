package template

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v7"

	"github.com/devian2011/msgchute/internal/dto"
)

var (
	MissingRequiredParamErr = errors.New("missing required param")
)

type Generator struct {
	tplSet *pongo2.TemplateSet
}

func NewGenerator() (*Generator, error) {
	tplSet := pongo2.NewSet("generator_set", pongo2.DefaultLoader)

	err := tplSet.RegisterFilter("uppercase", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, error) {
		return pongo2.AsValue(strings.ToUpper(in.String())), nil
	})
	if err != nil {
		return nil, err
	}

	return &Generator{
		tplSet: tplSet,
	}, nil
}

func (g *Generator) GenerateString(
	tmpl string,
	messageParams map[string]*dto.MessageParam,
	tmplParams map[string]*dto.TemplateParam,
) (string, error) {
	pTmpl, err := g.tplSet.FromString(tmpl)
	if err != nil {
		return "", err
	}

	params, buildParamErr := g.buildParams(messageParams, tmplParams)
	if buildParamErr != nil {
		return "", buildParamErr
	}
	context := pongo2.Context(params)

	result, err := pTmpl.Execute(context)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (g *Generator) buildParams(
	messageParams map[string]*dto.MessageParam,
	tmplParams map[string]*dto.TemplateParam,
) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(tmplParams))
	missing := make([]string, 0, len(tmplParams))
	for pName, pValue := range tmplParams {
		if inPVal, exists := messageParams[pName]; exists {
			result[pName] = inPVal.Value
		} else {
			if len(pValue.Default) == 0 {
				missing = append(missing, pName)
			} else {
				result[pName] = pValue.Default
			}
		}
	}
	if len(missing) > 0 {
		return nil, errors.Join(
			MissingRequiredParamErr,
			fmt.Errorf("missing required params: %s", strings.Join(missing, ", ")))
	}

	return result, nil
}
