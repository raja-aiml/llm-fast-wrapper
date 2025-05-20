{{- define "llm-fast-wrapper.name" -}}
llm-fast-wrapper
{{- end -}}

{{- define "llm-fast-wrapper.fullname" -}}
{{ include "llm-fast-wrapper.name" . }}
{{- end -}}
