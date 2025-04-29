{{- define "crds.name" -}}
{{- printf "%s-crds" (include "helm.name" $) -}}
{{- end }}

{{- define "crds.annotations" -}}
"helm.sh/hook": "pre-install,pre-upgrade"
  {{- with $.Values.global.jobs.annotations }}
    {{- . | toYaml | nindent 0 }}
  {{- end }}
{{- end }}

{{- define "crds.component" -}}
crd-install-hook
{{- end }}

{{- define "crds.regexReplace" -}}
{{- printf "%s" ($ | base | trimSuffix ".yaml" | regexReplaceAll "[_.]" "-") -}}
{{- end }}
