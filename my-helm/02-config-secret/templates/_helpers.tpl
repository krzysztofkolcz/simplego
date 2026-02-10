{{- define "go-hello.name" -}}
go-hello
{{- end }}

{{- define "go-hello.fullname" -}}
{{ include "go-hello.name" . }}-{{ .Release.Name }}
{{- end }}
