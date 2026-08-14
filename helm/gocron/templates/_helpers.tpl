{{/*
Expand the name of the chart.
*/}}
{{- define "gocron.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "gocron.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "gocron.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gocron.labels" -}}
helm.sh/chart: {{ include "gocron.chart" . }}
{{ include "gocron.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gocron.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gocron.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "gocron.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "gocron.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image tag
*/}}
{{- define "gocron.imageTag" -}}
{{- .Values.image.tag | default .Chart.AppVersion }}
{{- end }}

{{/* Validate requirements for the managed Kubernetes deployment. */}}
{{- define "gocron.validateValues" -}}
{{- if not (has .Values.db.engine (list "mysql" "postgres")) -}}
{{- fail "db.engine must be mysql or postgres; SQLite is not supported by the Kubernetes chart" -}}
{{- end -}}
{{- if not .Values.db.host -}}
{{- fail "db.host is required" -}}
{{- end -}}
{{- $currentSecret := lookup "v1" "Secret" .Release.Namespace (include "gocron.fullname" .) -}}
{{- if and (not .Values.managed.existingSecret) (not .Values.db.password) (not $currentSecret) -}}
{{- fail "db.password or managed.existingSecret is required" -}}
{{- end -}}
{{- if lt (int .Values.replicaCount) 1 -}}
{{- fail "replicaCount must be at least 1" -}}
{{- end -}}
{{- if lt (int .Values.db.maxOpenConns) 2 -}}
{{- fail "db.maxOpenConns must be at least 2 for managed bootstrap locking" -}}
{{- end -}}
{{- if and .Values.autoscaling.enabled (gt (int .Values.autoscaling.minReplicas) (int .Values.autoscaling.maxReplicas)) -}}
{{- fail "autoscaling.minReplicas cannot exceed autoscaling.maxReplicas" -}}
{{- end -}}
{{- end }}
