{{/**}}
Print out basic config information inside yaml tags for use in .md files with a replace clause.
This means you should be able to run the same generate on a file with these tags.

It is expected that once you setup the initial template comments in an .md file subsequent calls to the same file
are idempodent.
{{**/}}
{{- define "docsReplaceBasic" }}
{{`{{- template "docsReplaceBasic" . }}`}}
{{`{{- with deleteCurrent -}}`}}
--->
```yaml
{{ parseToConfigYamlOpts . true true }}
```
<!---
{{`{{- end }}`}}
{{- end }}

{{- define "docsReplaceFull" }}
{{`{{- template "docsReplaceFull" . }}`}}
{{`{{- with deleteCurrent -}}`}}
--->
```yaml
{{ parseToConfigYamlOpts . true false }}
```
<!---
{{`{{- end }}`}}
{{- end }}

