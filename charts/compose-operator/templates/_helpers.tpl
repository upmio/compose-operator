{{/* vim: set filetype=mustache: */}}

{{- define "compose-operator.fullname" -}}
{{- include "common.names.fullname" . -}}
{{- end -}}

{{/*
Return the proper ui image name
*/}}
{{- define "compose-operator.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "compose-operator.imagePullSecrets" -}}
{{- include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) }}
{{- end -}}

{{/*
Create the name of the cluster role to use
*/}}
{{- define "compose-operator.clusterRoleName" -}}
{{- if .Values.rbac.create }}
{{- default (include "common.names.fullname" .) .Values.rbac.name }}
{{- else }}
{{- default "default" .Values.rbac.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the cluster role binding to use
*/}}
{{- define "compose-operator.clusterRoleBindingName" -}}
{{- if .Values.rbac.create }}
{{- default (include "common.names.fullname" .) .Values.rbac.name }}
{{- else }}
{{- default "default" .Values.rbac.name }}
{{- end }}
{{- end }}

{{/*
 Create the name of the service account to use
 */}}
{{- define "compose-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{ default (include "common.names.fullname" .) .Values.serviceAccount.name | trunc 63 | trimSuffix "-" }}
{{- else -}}
{{ default "default" .Values.serviceAccount.name | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the role to use
*/}}
{{- define "compose-operator.roleName" -}}
{{- if .Values.rbac.create }}
{{- default (include "common.names.fullname" .) .Values.rbac.name }}
{{- else }}
{{- default "default" .Values.rbac.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the role binding to use
*/}}
{{- define "compose-operator.roleBindingName" -}}
{{- if .Values.rbac.create }}
{{- default (include "common.names.fullname" .) .Values.rbac.name }}
{{- else }}
{{- default "default" .Values.rbac.name }}
{{- end }}
{{- end }}