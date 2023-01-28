package pkg

const tpl = `package {{ .PackageName }}

import (
	{{- range $ip, $package := .Packages }}
	"{{$package}}"
	{{- end}}
)

{{ $receiver :=.EntityName -}}
func (t *{{ $receiver -}}) Validator() error {
	{{- define "main" -}}
	{{- range $if, $field := .Fields -}}
	{{ $starType :=.GetStarType -}}
	{{ $shouldNil :=.ShouldValidateNil -}}
	{{- range $it, $tag := $field.Tags -}}
	{{- if and  (eq $tag.Operator "required") (eq $shouldNil true) -}}
	if t.{{ $field.Field }} == nil {
		return errors.New("{{ $field.Field }} 不能为nil ")
	}
	{{- end -}}
	{{- $exist :=.Check $tag.Operator -}}
	{{- if and (ne $tag.Operator "") (eq $exist true) -}}
	{{- $get := .GetExp $field.Field $starType $tag.Operator $tag.Value $field.RealType -}}
	{{- if (ne $get "") -}}
	if {{$get}} {
		return errors.New("{{.GeError $field.Field $tag.Operator $tag.Value}}")
	}
	{{- end -}}
	{{- end -}}
	{{- if (eq $tag.Operator "required") -}}
	{{- if (eq $field.RealType "struct") }}
	if err := t.{{- $field.Field -}}.Validator(); err != nil {
		return err
	}
	{{- end -}}
	{{- end }}
	{{end -}}
	{{ end -}}
	{{- end}}
	{{ template "main" . -}}
	{{ range $ic, $cf := .CustomFuncs -}}
	if err := t.{{$cf.Name}}(); err !=nil {
		return err
	}
	{{end -}}
	return nil
}
`
