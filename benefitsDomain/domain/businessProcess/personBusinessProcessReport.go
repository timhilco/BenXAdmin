package businessProcess

import (
	"bytes"
	"text/template"
)

func (d *PersonBusinessProcess) Report(rc *ResourceContext) string {
	dir := rc.environmentVariables.TemplateDirectory
	templateFile := dir + "personBusinessProcessTemplate.tmpl"
	buf := new(bytes.Buffer)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buf, d)
	if err != nil {
		panic(err)
	}
	s1 := buf.String()
	dd := d.BusinessProcessData
	s2 := dd.Report(rc)
	return s1 + "\n" + s2

}
