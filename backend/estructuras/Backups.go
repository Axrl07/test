package estructuras

import (
	"fmt"
	"main/global"
	"time"
)

type Backup struct {
	Journal_start int64
	Recovery      string
	Journals      []Journal // lista de journals
}

// agrega un journal al backup
func (back *Backup) AgregarJournal(js Journal) {
	back.Journals = append(back.Journals, js)
}

// obtener último journal
func (back *Backup) ObtenerUltimoJournal() *Journal {
	if len(back.Journals) == 0 {
		// no hay journals
		return nil
	}
	// hay journals
	return &back.Journals[len(back.Journals)-1]
}

// regresa una tabla en html
func (back *Backup) ReporteJournals() string {
	salida := `<table class="table table-bordered table-striped table-hover align-middle text-center">`
	salida += `<thead class="table-dark"><tr>`
	salida += `<th scope="col">Operación</th>`
	salida += `<th scope="col">Path</th>`
	salida += `<th scope="col">Contenido</th>`
	salida += `<th scope="col">Fecha</th>`
	salida += `</tr></thead><tbody>`

	for _, js := range back.Journals {
		operacion := global.BorrandoIlegibles(string(js.J_content.I_operation[:]))
		path := global.BorrandoIlegibles(string(js.J_content.I_path[:]))
		contenido := global.BorrandoIlegibles(string(js.J_content.I_content[:]))
		fecha := global.ObtenerFecha(time.Unix(int64(js.J_content.I_date), 0))

		salida += "<tr>"
		salida += fmt.Sprintf("<td>%s</td>", operacion)
		salida += fmt.Sprintf("<td>%s</td>", path)
		salida += fmt.Sprintf("<td>%s</td>", contenido)
		salida += fmt.Sprintf("<td>%s</td>", fecha)
		salida += "</tr>"
	}

	salida += "</tbody></table>"
	return salida
}

// manejo de los backups
var (
	// almacena todos los journals creados según el ID de la partición
	Backups map[string]Backup = make(map[string]Backup)
)

func (back *Backup) ActualizarBackup(idPart string) {
	Backups[idPart] = *back
}

func (back *Backup) ActualizarRecovery(comando string) {
	back.Recovery += comando
}

// obtener el backup recursivamente (si no existe lo crea)
func ObtenerBackup(idPart string) *Backup {
	backup, ok := Backups[idPart]
	// si no existe lo crea
	if !ok {
		back := Backup{}
		Backups[idPart] = back
		return ObtenerBackup(idPart)
	}
	//retorna el back guardado si existe
	return &backup
}
