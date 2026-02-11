# {{- if env.Getenv "LATEST_RELEASE_IMAGE_URL" "" -}}
FROM "{{ env.Getenv "LATEST_RELEASE_IMAGE_URL" }}"
#{{- end }}
