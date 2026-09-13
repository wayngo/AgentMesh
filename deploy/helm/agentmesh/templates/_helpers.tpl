{{- define "agentmesh.name" -}}agentmesh{{- end }}
{{- define "agentmesh.labels" }}app.kubernetes.io/name: {{ include "agentmesh.name" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}{{- end }}