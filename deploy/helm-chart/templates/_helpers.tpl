{{- define "connector.labels" -}}
app: {{ .Values.connector.name }}
release: {{ .Release.Name }}
{{- end -}}

{{- define "fabedgeOperator.labels" -}}
app: {{ .Values.operator.name }}
release: {{ .Release.Name }}
{{- end -}}
