package businessProcess

import (
	"bytes"
	"text/template"
)

func (d *BusinessProcessDefinition) Report(rc *ResourceContext) string {
	dir := rc.environmentVariables.TemplateDirectory
	templateFile := dir + "businessProcessTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, d)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
